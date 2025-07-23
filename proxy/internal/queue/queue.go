package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/atyronesmith/llama-metrics/proxy/internal/metrics"
)

// Request represents a queued request
type Request struct {
	ID        string
	Model     string
	Handler   func() error
	Submitted time.Time
	ctx       context.Context
	result    chan error
}

// Manager handles request queuing and processing
type Manager struct {
	queue       chan *Request
	maxSize     int
	maxWorkers  int
	metrics     *metrics.Collector
	workerPool  sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc

	// Queue statistics
	mu              sync.RWMutex
	totalQueued     int64
	totalProcessed  int64
	totalRejected   int64
	currentSize     int
	peakSize        int
	lastProcessed   time.Time
}

// NewManager creates a new queue manager
func NewManager(maxSize, maxWorkers int, m *metrics.Collector) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	qm := &Manager{
		queue:      make(chan *Request, maxSize),
		maxSize:    maxSize,
		maxWorkers: maxWorkers,
		metrics:    m,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		qm.workerPool.Add(1)
		go qm.worker(i)
	}

	// Start metrics updater
	go qm.metricsUpdater()

	return qm
}

// Submit adds a request to the queue
func (qm *Manager) Submit(ctx context.Context, model string, handler func() error) error {
	req := &Request{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Model:     model,
		Handler:   handler,
		Submitted: time.Now(),
		ctx:       ctx,
		result:    make(chan error, 1),
	}

	// Try to add to queue
	select {
	case qm.queue <- req:
		qm.updateQueueStats(true)
		// Wait for result
		select {
		case err := <-req.result:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	default:
		// Queue is full
		qm.updateRejectedStats()
		return fmt.Errorf("queue is full (size: %d)", qm.maxSize)
	}
}

// worker processes requests from the queue
func (qm *Manager) worker(id int) {
	defer qm.workerPool.Done()

	for {
		select {
		case <-qm.ctx.Done():
			return
		case req := <-qm.queue:
			qm.processRequest(req)
		}
	}
}

// processRequest handles a single request
func (qm *Manager) processRequest(req *Request) {
	// Update queue stats
	qm.updateQueueStats(false)

	// Record queue wait time
	waitTime := time.Since(req.Submitted)
	qm.metrics.RecordQueueWaitTime(req.Model, waitTime)

	// Check if request context is still valid
	select {
	case <-req.ctx.Done():
		req.result <- req.ctx.Err()
		return
	default:
	}

	// Execute the handler
	err := req.Handler()
	req.result <- err

	// Update processed stats
	qm.updateProcessedStats()
}

// updateQueueStats updates queue statistics
func (qm *Manager) updateQueueStats(added bool) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if added {
		qm.totalQueued++
		qm.currentSize++
		if qm.currentSize > qm.peakSize {
			qm.peakSize = qm.currentSize
		}
	} else {
		qm.currentSize--
	}

	// Update metrics
	qm.metrics.QueueSize.Set(float64(qm.currentSize))
}

// updateProcessedStats updates processing statistics
func (qm *Manager) updateProcessedStats() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.totalProcessed++
	qm.lastProcessed = time.Now()
}

// updateRejectedStats updates rejection statistics
func (qm *Manager) updateRejectedStats() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.totalRejected++
	qm.metrics.RecordError("unknown", "queue_full")
}

// metricsUpdater periodically updates queue metrics
func (qm *Manager) metricsUpdater() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastProcessed int64
	lastUpdate := time.Now()

	for {
		select {
		case <-qm.ctx.Done():
			return
		case <-ticker.C:
			qm.mu.RLock()
			processed := qm.totalProcessed
			qm.mu.RUnlock()

			// Calculate processing rate
			duration := time.Since(lastUpdate).Seconds()
			rate := float64(processed-lastProcessed) / duration

			qm.metrics.RecordQueueProcessingRate(rate)

			lastProcessed = processed
			lastUpdate = time.Now()
		}
	}
}

// GetStats returns current queue statistics
func (qm *Manager) GetStats() map[string]interface{} {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	return map[string]interface{}{
		"current_size":    qm.currentSize,
		"max_size":        qm.maxSize,
		"peak_size":       qm.peakSize,
		"total_queued":    qm.totalQueued,
		"total_processed": qm.totalProcessed,
		"total_rejected":  qm.totalRejected,
		"workers":         qm.maxWorkers,
	}
}

// Shutdown gracefully shuts down the queue manager
func (qm *Manager) Shutdown(timeout time.Duration) error {
	// Stop accepting new requests
	qm.cancel()

	// Wait for workers to finish or timeout
	done := make(chan struct{})
	go func() {
		qm.workerPool.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("shutdown timeout after %v", timeout)
	}
}