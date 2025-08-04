// Dashboard initialization and main functionality
document.addEventListener('DOMContentLoaded', function() {
    // Initialize Bootstrap tooltips
    var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'))
    var tooltipList = tooltipTriggerList.map(function (tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl)
    });

    // Initialize charts
    initializeCharts();

    // Initial connection and data loading
    connectWebSocket();
    loadTimeSeriesData();
    loadMetrics();
    loadAIStatus();

    // Add refresh button click handler
    document.getElementById('refresh-status').addEventListener('click', function() {
        // Add loading indicator
        const button = this;
        const originalContent = button.innerHTML;
        button.innerHTML = '<i class="bi bi-arrow-repeat"></i>';
        button.disabled = true;

        // Load new status
        loadAIStatus();

        // Reset button after 2 seconds
        setTimeout(() => {
            button.innerHTML = originalContent;
            button.disabled = false;
        }, 2000);
    });

    // Set up periodic data refresh
    setInterval(loadTimeSeriesData, 30000); // Refresh time series data every 30 seconds
    setInterval(loadAIStatus, 30000); // Refresh AI status every 30 seconds

    // Handle window resize
    window.addEventListener('resize', handleChartResize);
});