// WebSocket connection management
let socket = null;
let reconnectInterval = null;
let reconnectAttempts = 0;
const maxReconnectAttempts = 50;
const reconnectDelay = 3000; // 3 seconds

function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    socket = new WebSocket(`${protocol}//${window.location.host}/ws`);

    // Handle incoming WebSocket messages
    socket.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            updateDashboard(data);
        } catch (error) {
            console.error('Error parsing WebSocket message:', error);
        }
    };

    // WebSocket event handlers
    socket.onopen = function() {
        console.log('Connected to dashboard');
        reconnectAttempts = 0; // Reset reconnect attempts on successful connection

        // Clear any existing reconnect interval
        if (reconnectInterval) {
            clearInterval(reconnectInterval);
            reconnectInterval = null;
        }

        // WebSocket connected - dashboard already shows as Online
    };

    socket.onclose = function() {
        console.log('Disconnected from dashboard');

        // Reset Ollama status to Unknown
        const ollamaDiv = document.getElementById('ollama-status');
        const ollamaBadge = ollamaDiv.querySelector('.badge');
        const ollamaResponseTime = document.getElementById('ollama-response-time');
        ollamaBadge.classList.remove('bg-success', 'bg-danger', 'bg-warning');
        ollamaBadge.classList.add('bg-secondary');
        ollamaBadge.textContent = 'Unknown';
        ollamaResponseTime.textContent = '';

        const ollamaTooltip = bootstrap.Tooltip.getInstance(ollamaDiv);
        if (ollamaTooltip) {
            ollamaTooltip.setContent({ '.tooltip-inner': 'Connection lost. Unable to determine Ollama status.' });
        }

        // Reset Proxy status to Unknown
        const proxyDiv = document.getElementById('proxy-status');
        const proxyBadge = proxyDiv.querySelector('.badge');
        const proxyResponseTime = document.getElementById('proxy-response-time');
        proxyBadge.classList.remove('bg-success', 'bg-danger', 'bg-warning');
        proxyBadge.classList.add('bg-secondary');
        proxyBadge.textContent = 'Unknown';
        proxyResponseTime.textContent = '';

        const proxyTooltip = bootstrap.Tooltip.getInstance(proxyDiv);
        if (proxyTooltip) {
            proxyTooltip.setContent({ '.tooltip-inner': 'Connection lost. Unable to determine proxy status.' });
        }

        // Attempt to reconnect
        if (!reconnectInterval && reconnectAttempts < maxReconnectAttempts) {
            reconnectInterval = setInterval(function() {
                if (socket.readyState === WebSocket.CLOSED) {
                    reconnectAttempts++;
                    console.log(`Attempting to reconnect... (${reconnectAttempts}/${maxReconnectAttempts})`);
                    connectWebSocket();
                }
            }, reconnectDelay);
        }
    };

    socket.onerror = function(error) {
        console.error('WebSocket error:', error);
    };
}