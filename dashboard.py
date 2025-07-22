#!/usr/bin/env python3
"""
Ollama Monitoring Dashboard
Real-time web dashboard for Ollama LLM performance metrics
"""

import json
import time
import logging
import requests
from datetime import datetime, timedelta
from flask import Flask, render_template, jsonify, request
from flask_socketio import SocketIO, emit
import threading
from typing import Dict, Any, List, Tuple
import asyncio
import aiohttp

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class LLMStatusGenerator:
    """Generate human-readable status summaries using LLM"""
    
    def __init__(self, ollama_url='http://localhost:11434', model='phi3:mini'):
        self.ollama_url = ollama_url  # Direct to Ollama, bypassing proxy
        self.model = model
        self.session = requests.Session()
        self.last_status = "System initializing..."
        self.skip_generation = False  # Flag to skip generation during high load
        
    def generate_status(self, metrics: Dict[str, Any]) -> tuple[str, bool]:
        """Generate a human-readable status summary from metrics
        Returns: (status_text, is_ai_generated)
        """
        try:
            # Get summary metrics
            summary = metrics.get('summary', {})
            
            # Skip generation if system is under high load
            active_requests = summary.get('active_requests', 0)
            queue_size = summary.get('queue_size', 0)
            
            if active_requests > 5 or queue_size > 10:
                # System under load, return a simple status without LLM
                tokens_per_sec = summary.get('tokens_per_second', 0)
                avg_latency = summary.get('avg_latency', 0)
                status = f"High load: {active_requests} active requests, {queue_size} queued. {tokens_per_sec:.1f} tokens/s, {avg_latency:.2f}s avg latency"
                return (status, False)  # Not AI generated
            
            # Only generate new status every 30 seconds to reduce load
            current_time = time.time()
            if hasattr(self, 'last_generation_time'):
                if current_time - self.last_generation_time < 30:
                    return (self.last_status, True)  # Cached AI status
            
            # Prepare metrics context for LLM
            context = self._prepare_metrics_context(metrics)
            
            # Create prompt for status generation
            prompt = self._create_status_prompt(context)
            
            # Query the LLM with short timeout
            response = self._query_llm(prompt, timeout=3.0)
            
            if response:
                self.last_status = response
                self.last_generation_time = current_time
                return (response, True)  # Fresh AI generated
            else:
                return (self.last_status, True)  # Cached AI status
                
        except Exception as e:
            logger.error(f"Status generation failed: {e}")
            return (self.last_status, False)  # Error, return last status
    
    def _prepare_metrics_context(self, metrics: Dict[str, Any]) -> Dict[str, str]:
        """Convert raw metrics into contextual information"""
        summary = metrics.get('summary', {})
        percentiles = metrics.get('latency_percentiles', {})
        
        # Analyze trends and thresholds
        context = {
            'request_activity': self._analyze_request_activity(summary.get('request_rate', 0)),
            'latency_status': self._analyze_latency(summary.get('avg_latency', 0), percentiles.get('p95', 0)),
            'gpu_status': self._analyze_gpu_usage(summary.get('gpu_utilization', 0)),
            'power_status': self._analyze_power(summary.get('power_consumption', 0)),
            'memory_status': self._analyze_memory(summary.get('memory_usage', 0)),
            'success_status': self._analyze_success_rate(summary.get('success_rate', 0)),
            'token_generation': self._analyze_token_rate(summary.get('tokens_per_second', 0)),
            'active_requests': summary.get('active_requests', 0)
        }
        
        return context
    
    def _analyze_request_activity(self, rate: float) -> str:
        if rate > 1.0:
            return "high activity"
        elif rate > 0.1:
            return "moderate activity"
        elif rate > 0:
            return "low activity"
        else:
            return "idle"
    
    def _analyze_latency(self, avg: float, p95: float) -> str:
        if avg > 10:
            return "high latency"
        elif avg > 5:
            return "elevated latency"
        elif avg > 2:
            return "normal latency"
        else:
            return "low latency"
    
    def _analyze_gpu_usage(self, usage: float) -> str:
        if usage > 50:
            return "high GPU usage"
        elif usage > 10:
            return "moderate GPU usage"
        elif usage > 1:
            return "light GPU usage"
        else:
            return "minimal GPU usage"
    
    def _analyze_power(self, power: float) -> str:
        if power > 25:
            return "high power consumption"
        elif power > 15:
            return "elevated power usage"
        elif power > 8:
            return "normal power usage"
        else:
            return "low power mode"
    
    def _analyze_memory(self, memory_mb: float) -> str:
        if memory_mb > 1000:
            return "high memory usage"
        elif memory_mb > 500:
            return "moderate memory usage"
        else:
            return "normal memory usage"
    
    def _analyze_success_rate(self, rate: float) -> str:
        if rate >= 99:
            return "excellent reliability"
        elif rate >= 95:
            return "good reliability"
        elif rate >= 90:
            return "fair reliability"
        else:
            return "reliability issues"
    
    def _analyze_token_rate(self, rate: float) -> str:
        if rate > 100:
            return "fast token generation"
        elif rate > 50:
            return "good token generation"
        elif rate > 10:
            return "normal token generation"
        elif rate > 0:
            return "slow token generation"
        else:
            return "no token generation"
    
    def _create_status_prompt(self, context: Dict[str, str]) -> str:
        """Create a prompt for LLM to generate status summary"""
        return f"""You are a system monitor for an AI model server. Create a detailed, human-readable status summary (2-4 sentences, max 500 characters) based on these conditions:

Request Activity: {context['request_activity']}
Latency: {context['latency_status']}
GPU: {context['gpu_status']}
Power: {context['power_status']}
Memory: {context['memory_status']}
Reliability: {context['success_status']}
Token Generation: {context['token_generation']}
Active Requests: {context['active_requests']}

Generate a comprehensive status paragraph like:
- "The LLM server is experiencing high activity with elevated GPU usage and power consumption at 28W. Response latency is within normal ranges at 2.1s average. Memory usage is stable at 850MB with excellent reliability. Token generation is performing well at 75 tokens/sec."
- "System is currently idle with minimal resource usage. GPU utilization is low at 2%, power consumption is efficient at 8W. The server is ready and optimized to handle incoming requests with fast response times."
- "Moderate request activity detected with 3 active connections. GPU is working at 15% capacity, power draw is normal at 18W. Average latency is good at 1.8s with strong token generation rates."

Status:"""
    
    def _query_llm(self, prompt: str, timeout: float = 5.0) -> str:
        """Query the LLM for status generation"""
        try:
            payload = {
                'model': self.model,
                'prompt': prompt,
                'stream': False,
                'options': {
                    'temperature': 0.3,
                    'max_tokens': 150,
                    'stop': ['\n', '---']
                }
            }
            
            response = self.session.post(
                f"{self.ollama_url}/api/generate",
                json=payload,
                timeout=timeout
            )
            
            if response.status_code == 200:
                data = response.json()
                status_text = data.get('response', '').strip()
                
                # Clean up the response - limit length but allow paragraph
                status_text = status_text.strip()
                if len(status_text) > 500:
                    status_text = status_text[:497] + "..."
                    
                return status_text
            else:
                logger.warning(f"LLM query failed: {response.status_code}")
                return ""
                
        except Exception as e:
            logger.error(f"LLM query error: {e}")
            return ""

class PrometheusClient:
    """Client for querying Prometheus metrics"""
    
    def __init__(self, prometheus_url='http://localhost:9090'):
        self.prometheus_url = prometheus_url
        self.session = requests.Session()
        
    def query(self, query: str) -> Dict[str, Any]:
        """Execute a Prometheus query"""
        try:
            url = f"{self.prometheus_url}/api/v1/query"
            params = {'query': query}
            
            response = self.session.get(url, params=params, timeout=10)
            response.raise_for_status()
            
            return response.json()
        except Exception as e:
            logger.error(f"Prometheus query failed: {query} - {e}")
            return {'status': 'error', 'data': {'result': []}}
    
    def query_range(self, query: str, start_time: str, end_time: str, step: str = '30s') -> Dict[str, Any]:
        """Execute a Prometheus range query"""
        try:
            url = f"{self.prometheus_url}/api/v1/query_range"
            params = {
                'query': query,
                'start': start_time,
                'end': end_time,
                'step': step
            }
            
            response = self.session.get(url, params=params, timeout=15)
            response.raise_for_status()
            
            return response.json()
        except Exception as e:
            logger.error(f"Prometheus range query failed: {query} - {e}")
            return {'status': 'error', 'data': {'result': []}}

class DashboardMetrics:
    """Collect and process dashboard metrics"""
    
    def __init__(self, prometheus_client):
        self.client = prometheus_client
        
    def get_summary_metrics(self) -> Dict[str, Any]:
        """Get high-level summary metrics"""
        metrics = {}
        
        # Request rate (requests per second)
        result = self.client.query('rate(ollama_proxy_requests_total[5m])')
        if result['status'] == 'success' and result['data']['result']:
            total_rps = sum(float(r['value'][1]) for r in result['data']['result'])
            metrics['request_rate'] = round(total_rps, 2)
        else:
            metrics['request_rate'] = 0.0
            
        # Average response time (focus on generate endpoint which is most important)
        result = self.client.query(
            'sum(rate(ollama_proxy_request_duration_seconds_sum{endpoint="/api/generate"}[5m])) / '
            'sum(rate(ollama_proxy_request_duration_seconds_count{endpoint="/api/generate"}[5m]))'
        )
        if result['status'] == 'success' and result['data']['result']:
            try:
                avg_latency = float(result['data']['result'][0]['value'][1])
                if str(avg_latency).lower() in ['nan', 'inf', '-inf']:
                    metrics['avg_latency'] = 0.0
                else:
                    metrics['avg_latency'] = round(avg_latency, 2)
            except (ValueError, TypeError):
                metrics['avg_latency'] = 0.0
        else:
            metrics['avg_latency'] = 0.0
            
        # Success rate
        success_result = self.client.query('rate(ollama_proxy_requests_total{status="200"}[5m])')
        total_result = self.client.query('rate(ollama_proxy_requests_total[5m])')
        
        success_rate = 0.0
        if (success_result['status'] == 'success' and success_result['data']['result'] and 
            total_result['status'] == 'success' and total_result['data']['result']):
            
            success_rps = sum(float(r['value'][1]) for r in success_result['data']['result'])
            total_rps = sum(float(r['value'][1]) for r in total_result['data']['result'])
            
            if total_rps > 0:
                success_rate = (success_rps / total_rps) * 100
                
        metrics['success_rate'] = round(success_rate, 1)
        
        # Token generation rate
        result = self.client.query('rate(ollama_proxy_generated_tokens_total[5m])')
        if result['status'] == 'success' and result['data']['result']:
            tokens_per_sec = sum(float(r['value'][1]) for r in result['data']['result'])
            metrics['tokens_per_second'] = round(tokens_per_sec, 1)
        else:
            metrics['tokens_per_second'] = 0.0
            
        # GPU utilization
        result = self.client.query('ollama_proxy_gpu_active_residency_percent')
        if result['status'] == 'success' and result['data']['result']:
            gpu_util = float(result['data']['result'][0]['value'][1])
            metrics['gpu_utilization'] = round(gpu_util, 1)
        else:
            metrics['gpu_utilization'] = 0.0
            
        # Power consumption
        result = self.client.query('ollama_proxy_package_power_watts')
        if result['status'] == 'success' and result['data']['result']:
            power = float(result['data']['result'][0]['value'][1])
            metrics['power_consumption'] = round(power, 1)
        else:
            metrics['power_consumption'] = 0.0
            
        # Memory usage
        result = self.client.query('ollama_proxy_memory_usage_bytes')
        if result['status'] == 'success' and result['data']['result']:
            memory_bytes = float(result['data']['result'][0]['value'][1])
            memory_mb = memory_bytes / (1024 * 1024)
            metrics['memory_usage'] = round(memory_mb, 1)
        else:
            metrics['memory_usage'] = 0.0
            
        # Active requests
        result = self.client.query('sum(ollama_proxy_active_requests)')
        if result['status'] == 'success' and result['data']['result']:
            active = float(result['data']['result'][0]['value'][1])
            metrics['active_requests'] = int(active)
        else:
            metrics['active_requests'] = 0
        
        # Queue metrics
        result = self.client.query('ollama_proxy_queue_size')
        if result['status'] == 'success' and result['data']['result']:
            queue_size = float(result['data']['result'][0]['value'][1])
            metrics['queue_size'] = int(queue_size)
        else:
            metrics['queue_size'] = 0
            
        result = self.client.query('ollama_proxy_queue_processing_rate')
        if result['status'] == 'success' and result['data']['result']:
            processing_rate = float(result['data']['result'][0]['value'][1])
            metrics['queue_processing_rate'] = round(processing_rate, 2)
        else:
            metrics['queue_processing_rate'] = 0.0
            
        result = self.client.query('ollama_proxy_max_queue_size')
        if result['status'] == 'success' and result['data']['result']:
            max_queue = float(result['data']['result'][0]['value'][1])
            metrics['max_queue_size'] = int(max_queue)
        else:
            metrics['max_queue_size'] = 0
            
        # Ollama health status
        metrics['ollama_status'] = self._check_ollama_health()
        
        return metrics
    
    def get_latency_percentiles(self) -> Dict[str, float]:
        """Get latency percentiles"""
        percentiles = {}
        
        for p in [50, 75, 95, 99]:
            quantile = p / 100.0
            query = f'histogram_quantile({quantile}, rate(ollama_proxy_request_duration_seconds_bucket[5m]))'
            result = self.client.query(query)
            
            if result['status'] == 'success' and result['data']['result']:
                try:
                    value = float(result['data']['result'][0]['value'][1])
                    if str(value).lower() in ['nan', 'inf', '-inf']:
                        percentiles[f'p{p}'] = 0.0
                    else:
                        percentiles[f'p{p}'] = round(value, 3)
                except (ValueError, TypeError):
                    percentiles[f'p{p}'] = 0.0
            else:
                percentiles[f'p{p}'] = 0.0
                
        return percentiles
    
    def _check_ollama_health(self) -> Dict[str, Any]:
        """Check Ollama server health status"""
        import requests
        import time
        
        status = {
            'status': 'unknown',
            'response_time': None,
            'last_check': time.time()
        }
        
        try:
            start_time = time.time()
            # Quick health check to Ollama API
            response = requests.get('http://localhost:11434/api/tags', timeout=5)
            response_time = (time.time() - start_time) * 1000  # Convert to milliseconds
            
            if response.status_code == 200:
                status['status'] = 'healthy'
                status['response_time'] = round(response_time, 1)
            else:
                status['status'] = 'unhealthy'
                status['response_time'] = round(response_time, 1)
                
        except requests.exceptions.Timeout:
            status['status'] = 'timeout'
        except requests.exceptions.ConnectionError:
            status['status'] = 'offline'
        except Exception as e:
            status['status'] = 'error'
            logger.error(f"Ollama health check failed: {e}")
        
        return status
    
    def get_time_series_data(self, hours: int = 1) -> Dict[str, List]:
        """Get time series data for graphs"""
        end_time = datetime.now()
        start_time = end_time - timedelta(hours=hours)
        
        start_timestamp = int(start_time.timestamp())
        end_timestamp = int(end_time.timestamp())
        
        data = {}
        
        # Token generation rate over time
        result = self.client.query_range(
            'rate(ollama_proxy_generated_tokens_total[1m])',
            str(start_timestamp), str(end_timestamp), '30s'
        )
        
        if result['status'] == 'success' and result['data']['result']:
            series = result['data']['result'][0]['values'] if result['data']['result'] else []
            data['tokens_per_second'] = [
                {'x': int(point[0]) * 1000, 'y': float(point[1])} 
                for point in series
            ]
        else:
            data['tokens_per_second'] = []
            
        # Memory usage over time
        result = self.client.query_range(
            'ollama_proxy_memory_usage_bytes / 1024 / 1024',
            str(start_timestamp), str(end_timestamp), '30s'
        )
        
        if result['status'] == 'success' and result['data']['result']:
            series = result['data']['result'][0]['values'] if result['data']['result'] else []
            data['memory_usage'] = [
                {'x': int(point[0]) * 1000, 'y': float(point[1])} 
                for point in series
            ]
        else:
            data['memory_usage'] = []
            
        # GPU utilization over time
        result = self.client.query_range(
            'ollama_proxy_gpu_active_residency_percent',
            str(start_timestamp), str(end_timestamp), '30s'
        )
        
        if result['status'] == 'success' and result['data']['result']:
            series = result['data']['result'][0]['values'] if result['data']['result'] else []
            data['gpu_utilization'] = [
                {'x': int(point[0]) * 1000, 'y': float(point[1])} 
                for point in series
            ]
        else:
            data['gpu_utilization'] = []
            
        # Power consumption over time
        result = self.client.query_range(
            'ollama_proxy_package_power_watts',
            str(start_timestamp), str(end_timestamp), '30s'
        )
        
        if result['status'] == 'success' and result['data']['result']:
            series = result['data']['result'][0]['values'] if result['data']['result'] else []
            data['power_consumption'] = [
                {'x': int(point[0]) * 1000, 'y': float(point[1])} 
                for point in series
            ]
        else:
            data['power_consumption'] = []
            
        # Queue size over time
        result = self.client.query_range(
            'ollama_proxy_queue_size',
            str(start_timestamp), str(end_timestamp), '30s'
        )
        
        if result['status'] == 'success' and result['data']['result']:
            series = result['data']['result'][0]['values'] if result['data']['result'] else []
            data['queue_size'] = [
                {'x': int(point[0]) * 1000, 'y': float(point[1])} 
                for point in series
            ]
        else:
            data['queue_size'] = []
            
        # Queue processing rate over time
        result = self.client.query_range(
            'ollama_proxy_queue_processing_rate',
            str(start_timestamp), str(end_timestamp), '30s'
        )
        
        if result['status'] == 'success' and result['data']['result']:
            series = result['data']['result'][0]['values'] if result['data']['result'] else []
            data['queue_processing_rate'] = [
                {'x': int(point[0]) * 1000, 'y': float(point[1])} 
                for point in series
            ]
        else:
            data['queue_processing_rate'] = []
            
        return data

# Initialize Flask app and components
app = Flask(__name__)
app.config['SECRET_KEY'] = 'ollama-dashboard-secret'
socketio = SocketIO(app, cors_allowed_origins="*")

prometheus_client = PrometheusClient()
metrics_collector = DashboardMetrics(prometheus_client)
status_generator = LLMStatusGenerator()

@app.route('/')
def dashboard():
    """Main dashboard page"""
    return render_template('dashboard.html')

@app.route('/api/metrics/summary')
def api_metrics_summary():
    """API endpoint for summary metrics"""
    metrics = metrics_collector.get_summary_metrics()
    percentiles = metrics_collector.get_latency_percentiles()
    
    return jsonify({
        'summary': metrics,
        'latency_percentiles': percentiles,
        'timestamp': datetime.now().isoformat()
    })

@app.route('/api/status')
def api_status():
    """API endpoint for AI-generated status summary"""
    try:
        metrics = metrics_collector.get_summary_metrics()
        percentiles = metrics_collector.get_latency_percentiles()
        
        metrics_data = {
            'summary': metrics,
            'latency_percentiles': percentiles
        }
        
        status, is_ai_generated = status_generator.generate_status(metrics_data)
        
        return jsonify({
            'status': status,
            'is_ai_generated': is_ai_generated,
            'timestamp': datetime.now().isoformat()
        })
    except Exception as e:
        logger.error(f"Status API error: {e}")
        return jsonify({
            'status': 'System monitoring active',
            'timestamp': datetime.now().isoformat()
        })

@app.route('/api/metrics/timeseries')
def api_metrics_timeseries():
    """API endpoint for time series data"""
    hours = request.args.get('hours', 1, type=int)
    data = metrics_collector.get_time_series_data(hours)
    
    return jsonify({
        'data': data,
        'timestamp': datetime.now().isoformat()
    })

def background_metrics_broadcast():
    """Background thread to broadcast metrics updates"""
    status_counter = 0
    while True:
        try:
            # Get latest metrics
            metrics = metrics_collector.get_summary_metrics()
            percentiles = metrics_collector.get_latency_percentiles()
            
            # Generate AI status every 15 seconds (every 3rd cycle)
            ai_status = None
            if status_counter % 3 == 0:
                try:
                    metrics_data = {
                        'summary': metrics,
                        'latency_percentiles': percentiles
                    }
                    ai_status, is_ai_generated = status_generator.generate_status(metrics_data)
                except Exception as e:
                    logger.error(f"Status generation error: {e}")
                    ai_status = None
                    is_ai_generated = False
            
            # Broadcast to all connected clients
            broadcast_data = {
                'summary': metrics,
                'latency_percentiles': percentiles,
                'timestamp': datetime.now().isoformat()
            }
            
            if ai_status:
                broadcast_data['ai_status'] = ai_status
                broadcast_data['is_ai_generated'] = is_ai_generated
                logger.info(f"Broadcasting AI status: {ai_status} (AI: {is_ai_generated})")
            
            socketio.emit('metrics_update', broadcast_data)
            
            status_counter += 1
            time.sleep(5)  # Update every 5 seconds
            
        except Exception as e:
            logger.error(f"Background metrics broadcast error: {e}")
            time.sleep(10)  # Wait longer on error

@socketio.on('connect')
def handle_connect():
    """Handle client connection"""
    logger.info('Client connected to dashboard')
    
    # Send initial metrics
    try:
        metrics = metrics_collector.get_summary_metrics()
        percentiles = metrics_collector.get_latency_percentiles()
        
        emit('metrics_update', {
            'summary': metrics,
            'latency_percentiles': percentiles,
            'timestamp': datetime.now().isoformat()
        })
    except Exception as e:
        logger.error(f"Error sending initial metrics: {e}")

@socketio.on('disconnect')
def handle_disconnect():
    """Handle client disconnection"""
    logger.info('Client disconnected from dashboard')

if __name__ == '__main__':
    # Start background metrics broadcast thread
    metrics_thread = threading.Thread(target=background_metrics_broadcast, daemon=True)
    metrics_thread.start()
    
    logger.info("🚀 Starting Ollama Dashboard")
    logger.info("📊 Dashboard available at http://localhost:3001")
    
    # Run the Flask-SocketIO app
    socketio.run(app, host='0.0.0.0', port=3001, debug=False)