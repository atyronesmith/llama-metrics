// Helper function to handle null/NaN values
function safeNumber(value, decimals = 2) {
    if (value === null || value === undefined || isNaN(value)) {
        return '--';
    }
    return value.toFixed(decimals);
}

// Update dashboard with WebSocket data
function updateDashboard(data) {
    if (data.summary) {
        updateMetrics(data);
    }

    if (data.ai_status !== undefined) {
        updateAIStatus(data.ai_status, data.is_ai_generated);
    }

    if (data.latency_percentiles) {
        updateLatencyPercentiles(data.latency_percentiles);
    }

    if (data.high_priority_percentiles) {
        updateHighPriorityPercentiles(data.high_priority_percentiles);
    }

    // Update last updated time
    document.getElementById('last-updated').textContent = new Date(data.timestamp).toLocaleTimeString();
}

// Update metric displays
function updateMetrics(data) {
    const summary = data.summary;
    const percentiles = data.latency_percentiles;

    // Update Ollama status
    if (summary.ollama_status) {
        const statusDiv = document.getElementById('ollama-status');
        const statusBadge = statusDiv.querySelector('.badge');
        const responseTime = document.getElementById('ollama-response-time');
        const tooltip = bootstrap.Tooltip.getInstance(statusDiv);

        // Update status badge
        statusBadge.classList.remove('bg-success', 'bg-danger', 'bg-warning', 'bg-secondary');

        let newTooltip = '';
        const uptimeText = summary.ollama_status.uptime ? `\nUptime: ${summary.ollama_status.uptime}` : '';
        switch(summary.ollama_status.status) {
            case 'healthy':
                statusBadge.classList.add('bg-success');
                statusBadge.textContent = 'Healthy';
                newTooltip = `Ollama is running normally. Model is loaded and ready to process requests.${uptimeText}`;
                break;
            case 'unhealthy':
                statusBadge.classList.add('bg-warning');
                statusBadge.textContent = 'Unhealthy';
                newTooltip = `Ollama is responding but experiencing issues. Performance may be degraded.${uptimeText}`;
                break;
            case 'offline':
            case 'timeout':
                statusBadge.classList.add('bg-danger');
                statusBadge.textContent = 'Offline';
                newTooltip = 'Ollama service is not responding. Check if the service is running on port 11434.';
                break;
            default:
                statusBadge.classList.add('bg-secondary');
                statusBadge.textContent = 'Unknown';
                newTooltip = 'Unable to determine Ollama status. Checking connection...';
        }

        // Update tooltip
        if (tooltip) {
            tooltip.setContent({ '.tooltip-inner': newTooltip });
        }

        // Update response time
        if (summary.ollama_status.response_time !== null) {
            responseTime.textContent = `${summary.ollama_status.response_time}ms`;
        } else {
            responseTime.textContent = '';
        }
    }

    // Update Proxy status
    if (summary.proxy_status) {
        const statusDiv = document.getElementById('proxy-status');
        const statusBadge = statusDiv.querySelector('.badge');
        const responseTime = document.getElementById('proxy-response-time');
        const tooltip = bootstrap.Tooltip.getInstance(statusDiv);

        // Update status badge
        statusBadge.classList.remove('bg-success', 'bg-danger', 'bg-warning', 'bg-secondary');

        let newTooltip = '';
        const proxyUptimeText = summary.proxy_status.uptime ? `\nUptime: ${summary.proxy_status.uptime}` : '';
        switch(summary.proxy_status.status) {
            case 'healthy':
                statusBadge.classList.add('bg-success');
                statusBadge.textContent = 'Healthy';
                newTooltip = `Proxy is intercepting requests and collecting metrics normally on port 11435.${proxyUptimeText}`;
                break;
            case 'unhealthy':
                statusBadge.classList.add('bg-warning');
                statusBadge.textContent = 'Unhealthy';
                newTooltip = `Proxy is running but may have issues. Check logs for errors.${proxyUptimeText}`;
                break;
            case 'offline':
            case 'timeout':
                statusBadge.classList.add('bg-danger');
                statusBadge.textContent = 'Offline';
                newTooltip = 'Proxy service is down. Metrics are not being collected. Run "make start-proxy" to restart.';
                break;
            default:
                statusBadge.classList.add('bg-secondary');
                statusBadge.textContent = 'Unknown';
                newTooltip = 'Unable to determine proxy status. Checking connection to metrics port 8001...';
        }

        // Update tooltip
        if (tooltip) {
            tooltip.setContent({ '.tooltip-inner': newTooltip });
        }

        // Update response time
        if (summary.proxy_status.response_time !== null) {
            responseTime.textContent = `${summary.proxy_status.response_time}ms`;
        } else {
            responseTime.textContent = '';
        }
    }

    // Update summary metrics
    document.getElementById('request-rate').textContent = safeNumber(summary.request_rate, 2);
    document.getElementById('avg-latency').textContent = safeNumber(summary.avg_latency, 2);
    document.getElementById('success-rate').textContent = safeNumber(summary.success_rate, 1) === '--' ? '--' : safeNumber(summary.success_rate, 1) + '%';
    document.getElementById('tokens-per-sec').textContent = safeNumber(summary.tokens_per_second, 1);

    document.getElementById('gpu-util').textContent = safeNumber(summary.gpu_utilization, 1) === '--' ? '--' : safeNumber(summary.gpu_utilization, 1) + '%';
    document.getElementById('power-usage').textContent = safeNumber(summary.power_consumption, 1) === '--' ? '--' : safeNumber(summary.power_consumption, 1) + 'W';

    // Debug memory value
    console.log('Memory usage raw value:', summary.memory_usage);
    console.log('Memory usage formatted:', safeNumber(summary.memory_usage, 1));

    // Format memory with appropriate unit
    if (summary.memory_usage !== null && summary.memory_usage !== undefined && !isNaN(summary.memory_usage)) {
        if (summary.memory_usage >= 1024) {
            // Convert to GB if over 1024 MB
            document.getElementById('memory-usage').textContent = safeNumber(summary.memory_usage / 1024, 2) + ' GB';
        } else {
            document.getElementById('memory-usage').textContent = safeNumber(summary.memory_usage, 1) + ' MB';
        }
    } else {
        document.getElementById('memory-usage').textContent = '-- MB';
    }
    document.getElementById('active-requests').textContent = summary.active_requests || 0;

    // Update queue metrics
    document.getElementById('queue-size').textContent = summary.queue_size || 0;
    document.getElementById('queue-processing-rate').textContent = safeNumber(summary.queue_processing_rate, 2);
    document.getElementById('max-queue-size').textContent = summary.max_queue_size || 0;

    // Calculate queue efficiency (processed vs queued)
    const queueEfficiency = summary.queue_size > 0 ?
        Math.max(0, Math.min(100, (summary.queue_processing_rate * 100 / Math.max(1, summary.queue_size)))) : 100;
    const efficiencyFormatted = safeNumber(queueEfficiency, 0);
    document.getElementById('queue-efficiency').textContent = efficiencyFormatted === '--' ? '--' : efficiencyFormatted + '%';

    // Update combined priority queue metrics (Requests/Success format)
    const highPriorityTotal = summary.high_priority_items_total || 0;
    const highPrioritySuccess = summary.high_priority_success_total || 0;
    const normalPriorityTotal = summary.normal_priority_items_total || 0;
    const normalPrioritySuccess = summary.normal_priority_success_total || 0;
    
    document.getElementById('high-priority-combined').textContent = `${highPriorityTotal}/${highPrioritySuccess}`;
    document.getElementById('normal-priority-combined').textContent = `${normalPriorityTotal}/${normalPrioritySuccess}`;

    // Update percentiles
    document.getElementById('p50').textContent = safeNumber(percentiles.p50, 2) === '--' ? '--' : safeNumber(percentiles.p50, 2) + 's';
    document.getElementById('p75').textContent = safeNumber(percentiles.p75, 2) === '--' ? '--' : safeNumber(percentiles.p75, 2) + 's';
    document.getElementById('p95').textContent = safeNumber(percentiles.p95, 2) === '--' ? '--' : safeNumber(percentiles.p95, 2) + 's';
    document.getElementById('p99').textContent = safeNumber(percentiles.p99, 2) === '--' ? '--' : safeNumber(percentiles.p99, 2) + 's';

    // Update high priority percentiles
    document.getElementById('hp_p50').textContent = safeNumber(percentiles.hp_p50, 2) === '--' ? '--' : safeNumber(percentiles.hp_p50, 2) + 's';
    document.getElementById('hp_p75').textContent = safeNumber(percentiles.hp_p75, 2) === '--' ? '--' : safeNumber(percentiles.hp_p75, 2) + 's';
    document.getElementById('hp_p95').textContent = safeNumber(percentiles.hp_p95, 2) === '--' ? '--' : safeNumber(percentiles.hp_p95, 2) + 's';
    document.getElementById('hp_p99').textContent = safeNumber(percentiles.hp_p99, 2) === '--' ? '--' : safeNumber(percentiles.hp_p99, 2) + 's';

    // Update last updated time
    document.getElementById('last-updated').textContent = new Date(data.timestamp).toLocaleTimeString();
}

function updateLatencyPercentiles(percentiles) {
    if (percentiles) {
        Object.entries(percentiles).forEach(([key, value]) => {
            const element = document.getElementById(key);
            if (element) {
                element.textContent = value !== null ? value.toFixed(2) + 's' : '--';
            }
        });
    }
}

function updateHighPriorityPercentiles(percentiles) {
    if (percentiles) {
        Object.entries(percentiles).forEach(([key, value]) => {
            const element = document.getElementById('hp_' + key);
            if (element) {
                element.textContent = value !== null ? value.toFixed(2) + 's' : '--';
            }
        });
    }
}

function updateAIStatus(status, isAIGenerated) {
    console.log('Received AI status:', status, 'AI Generated:', isAIGenerated);
    document.getElementById('ai-status').textContent = status;
    document.getElementById('status-timestamp').innerHTML =
        'Updated at ' + new Date().toLocaleTimeString();

    // Update status mode badge
    const statusModeBadge = document.getElementById('status-mode');
    if (isAIGenerated === false) {
        statusModeBadge.textContent = 'High Load Mode';
        statusModeBadge.classList.remove('bg-success', 'bg-secondary');
        statusModeBadge.classList.add('bg-warning');
    } else {
        statusModeBadge.textContent = 'AI Analysis';
        statusModeBadge.classList.remove('bg-warning', 'bg-secondary');
        statusModeBadge.classList.add('bg-success');
    }
}

function loadMetrics() {
    fetch('/api/metrics/summary')
        .then(response => response.json())
        .then(data => {
            if (data.summary) {
                updateMetrics(data);
            }
            if (data.latency_percentiles) {
                updateLatencyPercentiles(data.latency_percentiles);
            }
            if (data.high_priority_percentiles) {
                updateHighPriorityPercentiles(data.high_priority_percentiles);
            }
        })
        .catch(error => {
            console.error('Error loading metrics:', error);
        });
}

// Load initial AI status
function loadAIStatus() {
    fetch('/api/status')
        .then(response => response.json())
        .then(data => {
            if (data.status) {
                document.getElementById('ai-status').textContent = data.status;
                document.getElementById('status-timestamp').innerHTML =
                    'Updated at ' + new Date().toLocaleTimeString();
                console.log('AI status loaded:', data.status, 'AI Generated:', data.is_ai_generated);

                // Update status mode badge
                const statusModeBadge = document.getElementById('status-mode');
                if (data.is_ai_generated === false) {
                    statusModeBadge.textContent = 'High Load Mode';
                    statusModeBadge.classList.remove('bg-success', 'bg-secondary');
                    statusModeBadge.classList.add('bg-warning');
                } else {
                    statusModeBadge.textContent = 'AI Analysis';
                    statusModeBadge.classList.remove('bg-warning', 'bg-secondary');
                    statusModeBadge.classList.add('bg-success');
                }
            }
        })
        .catch(error => {
            console.error('Error loading AI status:', error);
        });
}