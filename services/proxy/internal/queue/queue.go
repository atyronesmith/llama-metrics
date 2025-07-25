package queue

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/atyronesmith/llama-metrics/proxy/internal/metrics"
)

// Priority levels
const (
	PriorityNormal = 0
	PriorityHigh   = 1
)

// Request represents a queued request
type Request struct {
	ID        string
	Model     string
	Priority  int
	Handler   func() error
	Submitted time.Time
	ctx       context.Context
	result    chan error
}

// PriorityQueue implements heap.Interface for priority queuing
type PriorityQueue []*Request

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// Higher priority first
	if pq[i].Priority != pq[j].Priority {
		return pq[i].Priority > pq[j].Priority
	}
	// For same priority, earlier submission time first (FIFO)
	return pq[i].Submitted.Before(pq[j].Submitted)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Request)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// Manager handles request queuing and processing with priority
type Manager struct {
	pq          PriorityQueue
	pqMutex     sync.Mutex
	maxSize     int
	maxWorkers  int
	metrics     *metrics.Collector
	workerPool  sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	workSignal  chan struct{}

	// Queue statistics
	mu               sync.RWMutex
	totalQueued      int64
	totalProcessed   int64
	totalRejected    int64
	currentSize      int
	peakSize         int
	lastProcessed    time.Time
	highPriorityCount int
	normalPriorityCount int
}

// NewManager creates a new queue manager with priority support
func NewManager(maxSize, maxWorkers int, m *metrics.Collector) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	qm := &Manager{
		pq:         make(PriorityQueue, 0, maxSize),
		maxSize:    maxSize,
		maxWorkers: maxWorkers,
		metrics:    m,
		ctx:        ctx,
		cancel:     cancel,
		workSignal: make(chan struct{}, maxSize),
	}

	// Initialize the priority queue
	heap.Init(&qm.pq)

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		qm.workerPool.Add(1)
		go qm.worker(i)
	}

	// Start metrics updater
	go qm.metricsUpdater()

	return qm
}

// Submit adds a request to the queue with a priority
func (qm *Manager) Submit(ctx context.Context, model string, priority int, handler func() error) error {
	req := &Request{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Model:     model,
		Priority:  priority,
		Handler:   handler,
		Submitted: time.Now(),
		ctx:       ctx,
		result:    make(chan error, 1),
	}

	// Add to priority queue
	qm.pqMutex.Lock()
	if len(qm.pq) >= qm.maxSize {
		qm.pqMutex.Unlock()
		qm.updateRejectedStats()
		return fmt.Errorf("queue is full (size: %d)", qm.maxSize)
	}

	heap.Push(&qm.pq, req)
	qm.updateQueueStatsLocked(true, priority)
	qm.pqMutex.Unlock()

	// Signal workers
	select {
	case qm.workSignal <- struct{}{}:
	default:
	}

	// Wait for result
	select {
	case err := <-req.result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// worker processes requests from the priority queue
func (qm *Manager) worker(id int) {
	defer qm.workerPool.Done()

	for {
		select {
		case <-qm.ctx.Done():
			return
		case <-qm.workSignal:
			// Get next request from priority queue
			qm.pqMutex.Lock()
			if len(qm.pq) == 0 {
				qm.pqMutex.Unlock()
				continue
			}
			req := heap.Pop(&qm.pq).(*Request)
			qm.updateQueueStatsLocked(false, req.Priority)
			qm.pqMutex.Unlock()

			qm.processRequest(req)
		}
	}
}

// processRequest handles a single request
func (qm *Manager) processRequest(req *Request) {
	// Record queue wait time
	waitTime := time.Since(req.Submitted)
	qm.metrics.RecordQueueWaitTime(req.Model, waitTime)

	// Record priority-specific wait time
	if req.Priority == PriorityHigh {
		qm.metrics.QueueHighPriorityWaitTime.Observe(waitTime.Seconds())
	} else {
		qm.metrics.QueueNormalPriorityWaitTime.Observe(waitTime.Seconds())
	}

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

// updateQueueStatsLocked updates queue statistics (must be called with pqMutex locked)
func (qm *Manager) updateQueueStatsLocked(added bool, priority int) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if added {
		qm.totalQueued++
		qm.currentSize++
		if qm.currentSize > qm.peakSize {
			qm.peakSize = qm.currentSize
		}
		if priority == PriorityHigh {
			qm.highPriorityCount++
		} else {
			qm.normalPriorityCount++
		}
	} else {
		qm.currentSize--
		if priority == PriorityHigh {
			qm.highPriorityCount--
		} else {
			qm.normalPriorityCount--
		}
	}

	// Update metrics
	qm.metrics.QueueSize.Set(float64(qm.currentSize))
	qm.metrics.QueueHighPriorityCount.Set(float64(qm.highPriorityCount))
	qm.metrics.QueueNormalPriorityCount.Set(float64(qm.normalPriorityCount))
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
		"current_size":       qm.currentSize,
		"max_size":           qm.maxSize,
		"peak_size":          qm.peakSize,
		"total_queued":       qm.totalQueued,
		"total_processed":    qm.totalProcessed,
		"total_rejected":     qm.totalRejected,
		"workers":            qm.maxWorkers,
		"high_priority":      qm.highPriorityCount,
		"normal_priority":    qm.normalPriorityCount,
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