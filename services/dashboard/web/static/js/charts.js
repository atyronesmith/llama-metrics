// Chart configurations
const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    interaction: {
        intersect: false,
        mode: 'index'
    },
    scales: {
        x: {
            type: 'time',
            time: {
                displayFormats: {
                    minute: 'HH:mm',
                    hour: 'HH:mm'
                }
            },
            grid: {
                display: true,
                color: 'rgba(0,0,0,0.1)'
            }
        },
        y: {
            beginAtZero: true,
            grid: {
                display: true,
                color: 'rgba(0,0,0,0.1)'
            }
        }
    },
    plugins: {
        legend: {
            display: false
        }
    }
};

// Initialize charts
let tokensChart, memoryChart, gpuChart, powerChart, queueSizeChart, queueRateChart;

function initializeCharts() {
    tokensChart = new Chart(document.getElementById('tokensChart'), {
        type: 'line',
        data: {
            datasets: [{
                label: 'Tokens/sec',
                data: [],
                borderColor: '#ffc107',
                backgroundColor: 'rgba(255, 193, 7, 0.1)',
                fill: true,
                tension: 0.1
            }]
        },
        options: chartOptions
    });

    memoryChart = new Chart(document.getElementById('memoryChart'), {
        type: 'line',
        data: {
            datasets: [{
                label: 'Memory (MB)',
                data: [],
                borderColor: '#17a2b8',
                backgroundColor: 'rgba(23, 162, 184, 0.1)',
                fill: true,
                tension: 0.1
            }]
        },
        options: chartOptions
    });

    gpuChart = new Chart(document.getElementById('gpuChart'), {
        type: 'line',
        data: {
            datasets: [{
                label: 'GPU Utilization (%)',
                data: [],
                borderColor: '#dc3545',
                backgroundColor: 'rgba(220, 53, 69, 0.1)',
                fill: true,
                tension: 0.1
            }]
        },
        options: chartOptions
    });

    powerChart = new Chart(document.getElementById('powerChart'), {
        type: 'line',
        data: {
            datasets: [{
                label: 'Power (W)',
                data: [],
                borderColor: '#ffc107',
                backgroundColor: 'rgba(255, 193, 7, 0.1)',
                fill: true,
                tension: 0.1
            }]
        },
        options: chartOptions
    });

    // Queue charts
    queueSizeChart = new Chart(document.getElementById('queueSizeChart'), {
        type: 'line',
        data: {
            datasets: [{
                label: 'Queue Size',
                data: [],
                borderColor: '#dc3545',
                backgroundColor: 'rgba(220, 53, 69, 0.1)',
                fill: true,
                tension: 0.1,
                stepped: true
            }]
        },
        options: chartOptions
    });

    queueRateChart = new Chart(document.getElementById('queueRateChart'), {
        type: 'line',
        data: {
            datasets: [{
                label: 'Processing Rate (req/s)',
                data: [],
                borderColor: '#17a2b8',
                backgroundColor: 'rgba(23, 162, 184, 0.1)',
                fill: true,
                tension: 0.1
            }]
        },
        options: chartOptions
    });
}

// Load time series data
function loadTimeSeriesData() {
    fetch('/api/metrics/timeseries')
        .then(response => response.json())
        .then(data => {
            const series = data.data;

            // Update token chart
            tokensChart.data.datasets[0].data = series.tokens_per_second || [];
            tokensChart.update('none');

            // Update memory chart
            memoryChart.data.datasets[0].data = series.memory_usage || [];
            memoryChart.update('none');

            // Update GPU chart
            gpuChart.data.datasets[0].data = series.gpu_utilization || [];
            gpuChart.update('none');

            // Update power chart
            powerChart.data.datasets[0].data = series.power_consumption || [];
            powerChart.update('none');

            // Update queue charts
            queueSizeChart.data.datasets[0].data = series.queue_size || [];
            queueSizeChart.update('none');

            queueRateChart.data.datasets[0].data = series.queue_processing_rate || [];
            queueRateChart.update('none');
        })
        .catch(error => {
            console.error('Error loading time series data:', error);
        });
}

// Handle window resize
function handleChartResize() {
    tokensChart.resize();
    memoryChart.resize();
    gpuChart.resize();
    powerChart.resize();
    queueSizeChart.resize();
    queueRateChart.resize();
}