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
    
    def __init__(self, portkey_url='http://localhost:8787', model='phi3:mini'):
        self.portkey_url = portkey_url  # Portkey AI Gateway
        self.model = model
        self.session = requests.Session()
        self.last_status = "System operational"
        self.skip_generation = False  # Flag to skip generation during high load
        self.request_in_progress = False  # Prevent concurrent requests
        self.last_request_time = 0
        self.consecutive_timeouts = 0
        self.last_generation_time = 0  # Initialize to ensure it exists
        
    def generate_status(self, metrics: Dict[str, Any]) -> tuple[str, bool]:
        """Generate a human-readable status summary from metrics
        Returns: (status_text, is_ai_generated)
        """
        try:
            # Get summary metrics
            summary = metrics.get('summary', {})
            current_time = time.time()
            
            # Skip generation if system is under high load
            active_requests = summary.get('active_requests', 0)
            queue_size = summary.get('queue_size', 0)
            
            # Debug log to see what's triggering high load mode
            if active_requests > 0 or queue_size > 0:
                logger.debug(f"Load check: {active_requests} active requests, {queue_size} queued")
            
            # Only trigger high load mode if there are actually requests being processed
            # Since we're using LiteLLM now, we should check for actual request activity
            recent_requests = summary.get('requests_total', 0) - summary.get('previous_requests_total', 0)
            
            if active_requests > 5 or queue_size > 10 or recent_requests > 10:
                # System under load, return a simple status without LLM
                tokens_per_sec = summary.get('tokens_per_second', 0)
                avg_latency = summary.get('avg_latency', 0)
                status = f"High load: {active_requests} active requests, {queue_size} queued. {tokens_per_sec:.1f} tokens/s, {avg_latency:.2f}s avg latency"
                self.consecutive_timeouts = 0  # Reset timeout counter
                logger.info(f"Entering high load mode: {active_requests} active, {queue_size} queued")
                return (status, False)  # Not AI generated
            
            # Check if a request is already in progress
            if self.request_in_progress:
                # Return a timeout status if request has been pending too long
                if current_time - self.last_request_time > 10:
                    self.request_in_progress = False  # Reset flag
                    return ("⏱️ Status generation timed out - using cached status", False)
                else:
                    return (self.last_status, True)  # Still waiting
            
            # Only generate new status every 15 seconds
            if hasattr(self, 'last_generation_time'):
                if current_time - self.last_generation_time < 15:
                    return (self.last_status, True)  # Cached AI status
            
            # If we've had too many consecutive timeouts, wait longer
            if self.consecutive_timeouts >= 3:
                if current_time - self.last_generation_time < 60:
                    return (f"⚠️ LLM temporarily unavailable (retrying in {60 - int(current_time - self.last_generation_time)}s) - {self.last_status}", False)
            
            # Mark request as in progress
            self.request_in_progress = True
            self.last_request_time = current_time
            
            # Prepare metrics context for LLM
            context = self._prepare_metrics_context(metrics)
            
            # Create prompt for status generation
            prompt = self._create_status_prompt(context)
            
            # Query the LLM with short timeout
            response = self._query_llm(prompt, timeout=10.0)
            
            # Clear in-progress flag
            self.request_in_progress = False
            
            if response:
                self.last_status = response
                self.last_generation_time = current_time
                self.consecutive_timeouts = 0  # Reset timeout counter
                return (response, True)  # Fresh AI generated
            else:
                self.consecutive_timeouts += 1
                # Generate a basic status from metrics if no valid LLM response
                tokens_per_sec = summary.get('tokens_per_second', 0)
                avg_latency = summary.get('avg_latency', 0)
                active = summary.get('active_requests', 0)
                gpu_util = summary.get('gpu_utilization', 0)
                self.last_status = f"System operational: {active} active requests, {tokens_per_sec:.1f} tokens/s, {avg_latency:.2f}s latency, GPU {gpu_util:.0f}%"
                return (self.last_status, False)  # Fallback status
                
        except Exception as e:
            logger.error(f"Status generation failed: {e}")
            self.request_in_progress = False  # Clear flag on error
            self.consecutive_timeouts += 1
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
    
    def _get_recent_activity_rate(self, window_seconds: int = 30) -> float:
        """Calculate request rate for recent activity window"""
        # For now, return 0 until we can properly integrate with metrics collector
        # This will be enhanced in a future update
        return 0.0
    
    def _analyze_request_activity(self, rate: float) -> str:
        # Also check for recent activity spikes (last 30 seconds)
        recent_activity = self._get_recent_activity_rate(30)
        current_rate = max(rate, recent_activity)
        
        if current_rate > 2.0:
            return "very high activity"
        elif current_rate > 1.0:
            return "high activity"
        elif current_rate > 0.2:
            return "moderate activity"
        elif current_rate > 0:
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
        return f"""Generate a brief status summary for an AI server monitoring dashboard. Use the metrics below to create one paragraph (2-3 sentences).

Current metrics:
- Request Activity: {context['request_activity']}
- Latency: {context['latency_status']}
- GPU: {context['gpu_status']}
- Power: {context['power_status']}
- Memory: {context['memory_status']}
- Reliability: {context['success_status']}
- Token Generation: {context['token_generation']}

Write a status summary:"""
    
    def _query_llm(self, prompt: str, timeout: float = 5.0) -> str:
        """Query the LLM via Portkey AI Gateway"""
        try:
            # Use OpenAI-compatible format for Portkey with Ollama provider
            payload = {
                'model': self.model,
                'messages': [{'role': 'user', 'content': prompt}],
                'stream': False,
                'temperature': 0.3,
                'max_tokens': 150
            }
            
            headers = {
                'x-portkey-provider': 'ollama',
                'x-portkey-base-url': 'http://localhost:11434',
                'Content-Type': 'application/json'
            }
            
            response = self.session.post(
                f"{self.portkey_url}/v1/chat/completions",
                json=payload,
                headers=headers,
                timeout=timeout
            )
            
            if response.status_code == 200:
                data = response.json()
                # Parse OpenAI format response
                if 'choices' in data and len(data['choices']) > 0:
                    status_text = data['choices'][0]['message']['content'].strip()
                else:
                    status_text = data.get('response', '').strip()
                
                # Validate response - check if it looks like a valid status
                if not status_text:
                    logger.error("Empty response from LLM")
                    return ""
                
                # Check if response looks like an error or unrelated content
                error_indicators = ['sorry', 'I need', 'dictionary', 'python', 'document', 'instruction', 'essay', 'comprehensive guide']
                if any(indicator in status_text.lower() for indicator in error_indicators):
                    logger.error(f"Invalid LLM response detected: {status_text[:100]}...")
                    return ""
                
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
        
        # Local request tracking for more accurate rate calculation
        self.request_history = []  # [(timestamp, total_requests), ...]
        self.max_history_size = 20  # Keep last 20 data points
        self.last_total_requests = 0
    
    def _update_request_history(self, total_requests: float):
        """Update the local request history for rate calculation"""
        import time
        current_time = time.time()
        
        # Add new data point
        self.request_history.append((current_time, total_requests))
        
        # Trim history to max size
        if len(self.request_history) > self.max_history_size:
            self.request_history = self.request_history[-self.max_history_size:]
    
    def _calculate_local_request_rate(self) -> float:
        """Calculate request rate from local history"""
        if len(self.request_history) < 2:
            return 0.0
        
        # Get the oldest and newest data points
        oldest_time, oldest_requests = self.request_history[0]
        newest_time, newest_requests = self.request_history[-1]
        
        # Calculate time difference in seconds
        time_diff = newest_time - oldest_time
        
        if time_diff <= 0:
            return 0.0
        
        # Calculate requests per second
        request_diff = newest_requests - oldest_requests
        return max(0.0, request_diff / time_diff)
        
    def get_summary_metrics(self) -> Dict[str, Any]:
        """Get high-level summary metrics"""
        metrics = {}
        
        # Request rate (requests per second) - multiple calculation methods
        metrics['request_rate'] = 0.0
        
        # First, get current total requests to update local tracking
        current_result = self.client.query('ollama_proxy_requests_total')
        current_total = 0.0
        if current_result['status'] == 'success' and current_result['data']['result']:
            current_total = sum(float(r['value'][1]) for r in current_result['data']['result'])
            
            # Update our local request history
            self._update_request_history(current_total)
        
        # Get Portkey vs Direct routing breakdown
        portkey_result = self.client.query('ollama_proxy_requests_total{routing="portkey"}')
        direct_result = self.client.query('ollama_proxy_requests_total{routing="direct"}')
        
        portkey_total = 0.0
        direct_total = 0.0
        
        if portkey_result['status'] == 'success' and portkey_result['data']['result']:
            portkey_total = sum(float(r['value'][1]) for r in portkey_result['data']['result'])
        
        if direct_result['status'] == 'success' and direct_result['data']['result']:
            direct_total = sum(float(r['value'][1]) for r in direct_result['data']['result'])
        
        metrics['portkey_requests'] = int(portkey_total)
        metrics['direct_requests'] = int(direct_total)
        metrics['routing_ratio'] = round(portkey_total / max(current_total, 1) * 100, 1) if current_total > 0 else 0
        
        # Method 1: Use local tracking (most reliable for sparse data)
        local_rate = self._calculate_local_request_rate()
        if local_rate > 0:
            metrics['request_rate'] = round(local_rate, 2)
        
        # Method 2: Try Prometheus rate calculation (for more established data)
        if metrics['request_rate'] == 0.0:
            result = self.client.query('rate(ollama_proxy_requests_total[2m])')
            if result['status'] == 'success' and result['data']['result']:
                total_rps = sum(float(r['value'][1]) for r in result['data']['result'])
                if total_rps > 0:
                    metrics['request_rate'] = round(total_rps, 2)
        
        # Method 3: Simple recent change detection
        if metrics['request_rate'] == 0.0 and current_total > self.last_total_requests:
            # If we have new requests since last check, show a minimal rate
            if len(self.request_history) >= 2:
                # Calculate rate over the last few data points
                recent_time_diff = self.request_history[-1][0] - self.request_history[-2][0]
                recent_req_diff = self.request_history[-1][1] - self.request_history[-2][1]
                if recent_time_diff > 0 and recent_req_diff > 0:
                    metrics['request_rate'] = round(recent_req_diff / recent_time_diff, 2)
        
        # Update last total for next comparison
        self.last_total_requests = current_total
            
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
            
        # Portkey Gateway metrics - query directly from Portkey API
        portkey_health = self._get_portkey_health()
        if portkey_health:
            metrics['active_requests'] = 0  # Portkey doesn't expose active requests directly
            metrics['queue_size'] = 0       # Set to 0 for now
            metrics['queue_processing_rate'] = 0.0
            metrics['max_queue_size'] = 0
            metrics['portkey_gateway_status'] = 'healthy'
        else:
            # Fallback values when Portkey is unreachable
            metrics['active_requests'] = 0
            metrics['queue_size'] = 0
            metrics['queue_processing_rate'] = 0.0
            metrics['max_queue_size'] = 0
            metrics['portkey_gateway_status'] = 'unhealthy'
            
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
            # Try a simple models endpoint to test if LiteLLM proxy is working
            headers = {'Authorization': 'Bearer sk-1234567890abcdef'}
            response = requests.get('http://localhost:8000/v1/models', headers=headers, timeout=10)
            response_time = (time.time() - start_time) * 1000  # Convert to milliseconds
            
            if response.status_code == 200:
                models_data = response.json()
                # If we can get models, the service is working
                if 'data' in models_data and len(models_data['data']) > 0:
                    status['status'] = 'healthy'
                    status['response_time'] = round(response_time, 1)
                else:
                    status['status'] = 'unhealthy'  
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
    
    def _get_portkey_health(self) -> bool:
        """Get Portkey Gateway health information"""
        import requests
        
        try:
            response = requests.get('http://localhost:8787/', timeout=5)
            
            if response.status_code == 200:
                return True
            else:
                logger.warning(f"Portkey health check failed with status {response.status_code}")
                return False
                
        except requests.exceptions.RequestException as e:
            logger.warning(f"Portkey health check failed: {e}")
            return False
    
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

# Health check endpoints
@app.route('/health')
def health():
    """Comprehensive dashboard health check"""
    try:
        from healthcheck import get_health
        import asyncio
        
        # Run async health check in event loop
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        health_data = loop.run_until_complete(get_health())
        loop.close()
        
        status_code = 200 if health_data['status'] == 'healthy' else 503
        return jsonify(health_data), status_code
    except Exception as e:
        logger.error(f"Dashboard health check failed: {e}")
        return jsonify({
            'status': 'unhealthy',
            'error': str(e),
            'timestamp': datetime.now().isoformat(),
            'service': 'dashboard'
        }), 503

@app.route('/health/simple')
def health_simple():
    """Simple dashboard health check"""
    try:
        from healthcheck import get_health_simple
        import asyncio
        
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        health_data = loop.run_until_complete(get_health_simple())
        loop.close()
        
        status_code = 200 if health_data['status'] == 'healthy' else 503
        return jsonify(health_data), status_code
    except Exception as e:
        return jsonify({
            'status': 'unhealthy',
            'error': str(e),
            'timestamp': datetime.now().isoformat()
        }), 503

@app.route('/ready')
def readiness():
    """Dashboard readiness check"""
    try:
        from healthcheck import get_readiness
        readiness_data = get_readiness()
        
        # Add dashboard-specific readiness checks
        readiness_data['components']['dashboard'] = 'ready'
        readiness_data['components']['websocket'] = 'ready' if socketio else 'failed'
        readiness_data['components']['metrics_collector'] = 'ready' if metrics_collector else 'failed'
        
        status_code = 200 if readiness_data['ready'] else 503
        return jsonify(readiness_data), status_code
    except Exception as e:
        return jsonify({
            'ready': False,
            'error': str(e),
            'timestamp': datetime.now().isoformat()
        }), 503

@app.route('/live')
def liveness():
    """Dashboard liveness check"""
    try:
        from healthcheck import get_liveness
        liveness_data = get_liveness()
        
        # Add dashboard-specific liveness info
        liveness_data['dashboard_active'] = True
        liveness_data['websocket_active'] = socketio is not None
        
        status_code = 200 if liveness_data['alive'] else 503
        return jsonify(liveness_data), status_code
    except Exception as e:
        return jsonify({
            'alive': False,
            'error': str(e),
            'timestamp': datetime.now().isoformat()
        }), 503

def background_metrics_broadcast():
    """Background thread to broadcast metrics updates"""
    last_status_attempt = 0
    last_health_check = 0
    cached_health_status = "unknown"
    
    while True:
        try:
            # Get latest metrics
            metrics = metrics_collector.get_summary_metrics()
            percentiles = metrics_collector.get_latency_percentiles()
            
            current_time = time.time()
            
            # Check system health every 30 seconds (less frequent than AI status)
            if current_time - last_health_check >= 30:
                try:
                    from healthcheck import get_health_simple
                    import asyncio
                    
                    # Run async health check
                    loop = asyncio.new_event_loop()
                    asyncio.set_event_loop(loop)
                    health_data = loop.run_until_complete(get_health_simple())
                    loop.close()
                    
                    cached_health_status = health_data.get('status', 'unknown')
                    last_health_check = current_time
                    
                except Exception as e:
                    logger.error(f"Health check error: {e}")
                    cached_health_status = "error"
            
            # Try to generate AI status every 30 seconds
            ai_status = None
            is_ai_generated = True
            
            if current_time - last_status_attempt >= 30:
                try:
                    metrics_data = {
                        'summary': metrics,
                        'latency_percentiles': percentiles
                    }
                    ai_status, is_ai_generated = status_generator.generate_status(metrics_data)
                    last_status_attempt = current_time
                except Exception as e:
                    logger.error(f"Status generation error: {e}")
                    ai_status = "Status generation error - check system logs"
                    is_ai_generated = False
            
            # Always broadcast current data with fresh timestamp
            broadcast_data = {
                'summary': metrics,
                'latency_percentiles': percentiles,
                'timestamp': datetime.now().isoformat(),
                'system_health': cached_health_status,
                'high_load_mode': False  # Default to false, will be set by AI status logic
            }
            
            if ai_status:
                broadcast_data['ai_status'] = ai_status
                broadcast_data['is_ai_generated'] = is_ai_generated
                
                # Check if this was a high load response (not AI generated due to load)
                if not is_ai_generated and "High load:" in ai_status:
                    broadcast_data['high_load_mode'] = True
                # Only log AI-generated statuses to reduce noise
                if is_ai_generated and not ai_status.startswith("⏱️") and not ai_status.startswith("⚠️"):
                    logger.info(f"Broadcasting AI status: {ai_status} (AI: {is_ai_generated})")
            
            socketio.emit('metrics_update', broadcast_data)
            
            time.sleep(5)  # Update metrics every 5 seconds
            
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