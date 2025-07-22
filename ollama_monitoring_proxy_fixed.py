#!/usr/bin/env python3
"""
Ollama Monitoring Proxy - Fixed Version
Intercepts requests to Ollama and collects Prometheus metrics
"""

import asyncio
import json
import time
import logging
import psutil
import subprocess
import platform
import re
from aiohttp import web
from prometheus_client import (
    Counter, Histogram, Gauge, Info, generate_latest
)
from healthcheck import get_health, get_health_simple, get_readiness, get_liveness

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class OllamaMonitoringProxy:
    def __init__(self, ollama_host='localhost', ollama_port=11434,
                 proxy_port=11435, metrics_port=8001, 
                 portkey_host='localhost', portkey_port=8787,
                 enable_portkey=False):
        self.ollama_url = f"http://{ollama_host}:{ollama_port}"
        self.portkey_url = f"http://{portkey_host}:{portkey_port}"
        self.proxy_port = proxy_port
        self.metrics_port = metrics_port
        self.enable_portkey = enable_portkey

        # Initialize metrics
        self._init_metrics()

        # Track queue metrics
        self.queue_size = 0
        self.request_queue = asyncio.Queue()
        self.queue_semaphore = asyncio.Semaphore(10)  # Limit concurrent requests
        self.request_timestamps = {}  # Track when requests were queued
        self.processed_requests_count = 0
        self.last_processing_rate_update = time.time()

    def _init_metrics(self):
        """Initialize Prometheus metrics"""
        # Request metrics
        self.request_count = Counter(
            'ollama_proxy_requests_total',
            'Total number of requests',
            ['method', 'endpoint', 'model', 'status', 'routing']
        )

        self.request_duration = Histogram(
            'ollama_proxy_request_duration_seconds',
            'Request duration in seconds',
            ['method', 'endpoint', 'model', 'routing'],
            buckets=(0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 1.5, 2.0, 2.5, 5.0, 7.5, 10.0, 15.0, 30.0, 45.0, 60.0, 90.0, 120.0, 180.0)
        )

        self.active_requests = Gauge(
            'ollama_proxy_active_requests',
            'Number of active requests',
            ['model']
        )

        # Token metrics
        self.prompt_tokens = Counter(
            'ollama_proxy_prompt_tokens_total',
            'Total prompt tokens processed',
            ['model']
        )

        self.generated_tokens = Counter(
            'ollama_proxy_generated_tokens_total',
            'Total tokens generated',
            ['model']
        )

        self.tokens_per_second = Histogram(
            'ollama_proxy_tokens_per_second',
            'Tokens generated per second',
            ['model'],
            buckets=(10, 50, 100, 200, 500, 1000, 2000)
        )

        # Performance metrics
        self.time_to_first_token = Histogram(
            'ollama_proxy_time_to_first_token_seconds',
            'Time to first token in seconds',
            ['model'],
            buckets=(0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0)
        )

        self.model_load_duration = Histogram(
            'ollama_proxy_model_load_duration_seconds',
            'Model load duration in seconds',
            ['model'],
            buckets=(0.1, 0.5, 1.0, 5.0, 10.0, 30.0, 60.0)
        )

        # Error tracking
        self.error_count = Counter(
            'ollama_proxy_errors_total',
            'Total number of errors',
            ['model', 'error_type']
        )

        # Per-endpoint latency breakdown
        self.generate_latency = Histogram(
            'ollama_proxy_generate_latency_seconds',
            'Latency for /api/generate endpoint',
            ['model', 'streaming'],
            buckets=(0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 2.0, 5.0, 10.0, 20.0, 30.0, 60.0, 120.0)
        )

        self.chat_latency = Histogram(
            'ollama_proxy_chat_latency_seconds', 
            'Latency for /api/chat endpoint',
            ['model', 'streaming'],
            buckets=(0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 2.0, 5.0, 10.0, 20.0, 30.0, 60.0, 120.0)
        )

        self.tags_latency = Histogram(
            'ollama_proxy_tags_latency_seconds',
            'Latency for /api/tags endpoint', 
            [],
            buckets=(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0)
        )

        self.show_latency = Histogram(
            'ollama_proxy_show_latency_seconds',
            'Latency for /api/show endpoint',
            ['model'],
            buckets=(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0)
        )

        # System metrics
        self.cpu_usage = Gauge(
            'ollama_proxy_cpu_usage_percent',
            'CPU usage percentage'
        )

        self.memory_usage = Gauge(
            'ollama_proxy_memory_usage_bytes',
            'Memory usage in bytes'
        )

        # Queue metrics
        self.queue_size_gauge = Gauge(
            'ollama_proxy_queue_size',
            'Current request queue size'
        )
        
        self.queue_wait_time = Histogram(
            'ollama_proxy_queue_wait_seconds',
            'Time requests spend waiting in queue',
            ['model'],
            buckets=(0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0)
        )
        
        self.queue_processing_rate = Gauge(
            'ollama_proxy_queue_processing_rate',
            'Requests processed per second from queue'
        )
        
        self.max_queue_size = Gauge(
            'ollama_proxy_max_queue_size',
            'Maximum queue size observed'
        )
        
        self.queue_rejections = Counter(
            'ollama_proxy_queue_rejections_total',
            'Total number of requests rejected due to full queue'
        )

        # macOS GPU and power metrics
        self.gpu_active_residency = Gauge(
            'ollama_proxy_gpu_active_residency_percent',
            'GPU active residency percentage'
        )

        self.gpu_power = Gauge(
            'ollama_proxy_gpu_power_watts',
            'GPU power consumption in watts'
        )

        self.cpu_power = Gauge(
            'ollama_proxy_cpu_power_watts', 
            'CPU power consumption in watts'
        )

        self.package_power = Gauge(
            'ollama_proxy_package_power_watts',
            'Total package power consumption in watts'
        )

        self.cpu_temperature = Gauge(
            'ollama_proxy_cpu_temperature_celsius',
            'CPU temperature in Celsius'
        )

        self.gpu_temperature = Gauge(
            'ollama_proxy_gpu_temperature_celsius', 
            'GPU temperature in Celsius'
        )

        self.thermal_pressure = Gauge(
            'ollama_proxy_thermal_pressure_percent',
            'System thermal pressure percentage'
        )

        # Context length tracking
        self.context_length = Histogram(
            'ollama_proxy_context_length',
            'Context length in tokens',
            ['model'],
            buckets=[128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768]
        )

    async def proxy_request(self, request):
        """Proxy requests to Ollama and collect metrics"""
        import aiohttp

        path = request.path_qs
        method = request.method
        model = "unknown"
        start_time = time.time()

        # Extract model from request body if present
        body_bytes = None
        body_json = None

        if request.can_read_body:
            try:
                body_bytes = await request.read()
                if body_bytes:
                    body_json = json.loads(body_bytes.decode('utf-8'))
                    model = body_json.get('model', 'unknown')
            except Exception as e:
                logger.error(f"Error reading request body: {e}")
                body_json = {}

        # Track request entry to queue
        request_id = id(request)  # Unique identifier for this request
        queue_entry_time = time.time()
        self.request_timestamps[request_id] = queue_entry_time
        
        # Update queue size before acquiring semaphore
        self.queue_size += 1
        self.queue_size_gauge.set(self.queue_size)
        
        # Update max queue size if needed
        if hasattr(self, '_max_queue_seen'):
            if self.queue_size > self._max_queue_seen:
                self._max_queue_seen = self.queue_size
                self.max_queue_size.set(self._max_queue_seen)
        else:
            self._max_queue_seen = self.queue_size
            self.max_queue_size.set(self._max_queue_seen)
        
        # Acquire semaphore to limit concurrent requests
        await self.queue_semaphore.acquire()

        try:
            # Request is now leaving the queue and starting processing
            processing_start_time = time.time()
            queue_wait_time = processing_start_time - queue_entry_time
            
            # Update queue metrics
            self.queue_size -= 1
            self.queue_size_gauge.set(self.queue_size)
            self.queue_wait_time.labels(model=model).observe(queue_wait_time)
            
            # Update active requests (now we're actually processing)
            self.active_requests.labels(model=model).inc()
            
            # Clean up request timestamp tracking
            self.request_timestamps.pop(request_id, None)
            # Create a new session for the proxy request
            async with aiohttp.ClientSession() as session:
                # Choose destination based on Portkey setting
                if self.enable_portkey:
                    url = f"{self.portkey_url}{path}"
                    logger.debug(f"Routing through Portkey: {url}")
                else:
                    url = f"{self.ollama_url}{path}"
                    logger.debug(f"Direct routing to Ollama: {url}")

                # Build the request kwargs
                kwargs = {
                    'method': method,
                    'url': url,
                    'timeout': aiohttp.ClientTimeout(total=300)
                }

                # Handle different content types
                content_type = request.headers.get('content-type', '')

                if body_json and 'application/json' in content_type:
                    # Use json parameter for JSON payloads
                    kwargs['json'] = body_json
                elif body_bytes:
                    # Use data for other payloads
                    kwargs['data'] = body_bytes
                    # Only forward content-type for non-JSON requests
                    kwargs['headers'] = {'content-type': content_type}

                # Make the request to Ollama
                async with session.request(**kwargs) as resp:
                    # Handle streaming responses
                    if body_json and body_json.get('stream', False):
                        return await self._handle_streaming_response(
                            resp, request, model, start_time, path
                        )
                    else:
                        # Non-streaming response
                        response_body = await resp.read()
                        response_json = {}

                        if response_body:
                            try:
                                response_json = json.loads(response_body.decode('utf-8'))
                            except:
                                pass

                        # Extract metrics from response
                        if response_json:
                            self._extract_response_metrics(response_json, model, start_time)

                        # Update request metrics
                        duration = time.time() - start_time
                        routing_type = "portkey" if self.enable_portkey else "direct"
                        self.request_duration.labels(
                            method=method, endpoint=path, model=model, routing=routing_type
                        ).observe(duration)

                        # Record per-endpoint latency
                        self._record_endpoint_latency(path, duration, model, body_json)

                        routing_type = "portkey" if self.enable_portkey else "direct"
                        self.request_count.labels(
                            method=method, endpoint=path, model=model,
                            status=resp.status, routing=routing_type
                        ).inc()

                        # Return response
                        return web.Response(
                            body=response_body,
                            status=resp.status,
                            headers={'content-type': resp.headers.get('content-type', 'application/json')}
                        )

        except asyncio.TimeoutError:
            self.error_count.labels(model=model, error_type='timeout').inc()
            return web.Response(status=504, text="Gateway Timeout")
        except Exception as e:
            logger.error(f"Proxy error: {e}")
            self.error_count.labels(model=model, error_type='proxy_error').inc()
            return web.Response(status=502, text=f"Bad Gateway: {str(e)}")
        finally:
            # Decrement active requests and release semaphore
            self.active_requests.labels(model=model).dec()
            self.queue_semaphore.release()
            
            # Increment processed requests counter for processing rate calculation
            self.processed_requests_count += 1
            
            # Clean up any remaining request timestamp tracking
            self.request_timestamps.pop(request_id, None)

    async def _handle_streaming_response(self, resp, request, model, start_time, path):
        """Handle streaming responses from Ollama"""
        response = web.StreamResponse(
            status=resp.status,
            headers={'content-type': 'application/x-ndjson'}
        )
        await response.prepare(request)

        first_token_time = None
        total_tokens = 0

        try:
            async for line in resp.content:
                if line:
                    await response.write(line)

                    # Try to parse metrics from streaming chunks
                    try:
                        chunk = json.loads(line.decode('utf-8').strip())

                        if first_token_time is None and chunk.get('response'):
                            first_token_time = time.time() - start_time
                            self.time_to_first_token.labels(model=model).observe(first_token_time)

                        # Extract final metrics from done chunk
                        if chunk.get('done'):
                            self._extract_response_metrics(chunk, model, start_time)
                    except:
                        pass

        except Exception as e:
            logger.error(f"Streaming error: {e}")
            self.error_count.labels(model=model, error_type='streaming_error').inc()

        # Update request metrics
        duration = time.time() - start_time
        routing_type = "portkey" if self.enable_portkey else "direct"
        self.request_duration.labels(
            method='POST', endpoint=path, model=model, routing=routing_type
        ).observe(duration)

        # Record per-endpoint latency (streaming)
        self._record_endpoint_latency(path, duration, model, {"stream": True})

        routing_type = "portkey" if self.enable_portkey else "direct"
        self.request_count.labels(
            method='POST', endpoint=path, model=model, status=resp.status, routing=routing_type
        ).inc()

        await response.write_eof()
        return response

    def _extract_response_metrics(self, response_json, model, start_time):
        """Extract metrics from Ollama response"""
        # Token metrics
        if 'prompt_eval_count' in response_json:
            prompt_tokens = response_json['prompt_eval_count']
            self.prompt_tokens.labels(model=model).inc(prompt_tokens)

        if 'eval_count' in response_json:
            generated_tokens = response_json['eval_count']
            self.generated_tokens.labels(model=model).inc(generated_tokens)

        # Performance metrics
        if 'eval_duration' in response_json and response_json.get('eval_count'):
            eval_duration_s = response_json['eval_duration'] / 1e9
            tokens = response_json['eval_count']
            if eval_duration_s > 0:
                tps = tokens / eval_duration_s
                self.tokens_per_second.labels(model=model).observe(tps)

        if 'load_duration' in response_json:
            load_duration_s = response_json['load_duration'] / 1e9
            self.model_load_duration.labels(model=model).observe(load_duration_s)

        # Context length
        if 'prompt_eval_count' in response_json:
            self.context_length.labels(model=model).observe(response_json['prompt_eval_count'])

    def _collect_macos_gpu_metrics(self):
        """Collect GPU metrics using non-privileged macOS methods"""
        if platform.system() != 'Darwin':
            return {}
        
        metrics = {}
        
        # Method 1: Monitor process activity for GPU usage estimation
        try:
            # Check for processes using significant CPU (indication of GPU work)
            import psutil
            ollama_processes = []
            for proc in psutil.process_iter(['pid', 'name', 'cpu_percent']):
                try:
                    if 'ollama' in proc.info['name'].lower():
                        ollama_processes.append(proc)
                except (psutil.NoSuchProcess, psutil.AccessDenied):
                    pass
            
            if ollama_processes:
                total_cpu = sum(proc.info['cpu_percent'] or 0 for proc in ollama_processes)
                # Estimate GPU usage based on ollama CPU usage (heuristic)
                # On Apple Silicon, high ollama CPU often correlates with GPU usage
                estimated_gpu = min(total_cpu * 1.2, 100.0)  # Rough estimation
                metrics['gpu_active_residency'] = estimated_gpu
                
        except Exception as e:
            logger.debug(f"Process-based GPU estimation failed: {e}")
        
        # Method 2: Use system_profiler for static GPU info
        try:
            result = subprocess.run([
                'system_profiler', 'SPDisplaysDataType', '-json'
            ], capture_output=True, text=True, timeout=5)
            
            if result.returncode == 0:
                import json
                try:
                    data = json.loads(result.stdout)
                    # This gives us static info, but we can at least confirm GPU is active
                    if data.get('SPDisplaysDataType'):
                        # GPU is present and active
                        if 'gpu_active_residency' not in metrics:
                            metrics['gpu_active_residency'] = 0.0  # Base activity level
                except json.JSONDecodeError:
                    pass
                        
        except Exception as e:
            logger.debug(f"system_profiler GPU collection failed: {e}")
            
        return metrics

    def _collect_macos_power_metrics(self):
        """Collect power metrics using non-privileged methods on macOS"""
        if platform.system() != 'Darwin':
            return {}
            
        metrics = {}
        
        # Method 1: Estimate power based on CPU utilization
        try:
            import psutil
            cpu_percent = psutil.cpu_percent(interval=0.1)
            
            # Rough power estimation for Apple Silicon Macs
            # M4 Pro typical power consumption ranges:
            # - Idle: 3-5W
            # - Medium load: 8-15W  
            # - High load: 20-35W
            base_power = 4.0  # Base idle power
            max_additional = 30.0  # Max additional power under load
            
            estimated_cpu_power = base_power + (cpu_percent / 100.0) * max_additional
            metrics['cpu_power'] = estimated_cpu_power
            
            # Estimate package power (CPU + GPU + system)
            metrics['package_power'] = estimated_cpu_power * 1.3  # Include GPU and system
            
        except Exception as e:
            logger.debug(f"Power estimation failed: {e}")
            
        # Method 2: Check system load for better estimation
        try:
            # Get system load average
            load_avg = psutil.getloadavg()[0]  # 1-minute load average
            cpu_count = psutil.cpu_count()
            
            if cpu_count > 0:
                load_ratio = min(load_avg / cpu_count, 1.0)
                # Adjust power estimation based on load
                if 'cpu_power' not in metrics:
                    base_power = 5.0
                    max_power = 35.0
                    metrics['cpu_power'] = base_power + (load_ratio * (max_power - base_power))
                    metrics['package_power'] = metrics['cpu_power'] * 1.4
                    
        except Exception as e:
            logger.debug(f"Load-based power estimation failed: {e}")
            
        return metrics

    def _collect_macos_temperature_metrics(self):
        """Collect temperature metrics on macOS using thermal monitoring"""
        if platform.system() != 'Darwin':
            return {}
            
        try:
            # Get thermal pressure information
            result = subprocess.run([
                'pmset', '-g', 'therm'
            ], capture_output=True, text=True, timeout=5)
            
            if result.returncode == 0:
                output = result.stdout
                metrics = {}
                
                # Parse thermal state
                if 'CPU_Scheduler_Limit' in output:
                    match = re.search(r'CPU_Scheduler_Limit\s*=\s*(\d+)', output)
                    if match:
                        # Convert to thermal pressure percentage (100 - scheduler limit)
                        scheduler_limit = int(match.group(1))
                        metrics['thermal_pressure'] = max(0, 100 - scheduler_limit)
                
                return metrics
                
        except (subprocess.TimeoutExpired, subprocess.CalledProcessError, Exception) as e:
            logger.debug(f"Temperature metrics collection failed: {e}")
            
        return {}

    async def collect_system_metrics(self):
        """Collect system metrics periodically"""
        while True:
            try:
                # CPU usage
                cpu_percent = psutil.cpu_percent(interval=1)
                self.cpu_usage.set(cpu_percent)

                # Memory usage
                memory = psutil.Process().memory_info()
                self.memory_usage.set(memory.rss)

                # Queue size
                self.queue_size_gauge.set(self.queue_size)
                
                # Calculate processing rate (requests per second)
                current_time = time.time()
                time_elapsed = current_time - self.last_processing_rate_update
                if time_elapsed >= 10:  # Update every 10 seconds
                    requests_processed_in_period = self.processed_requests_count
                    processing_rate = requests_processed_in_period / time_elapsed
                    self.queue_processing_rate.set(processing_rate)
                    
                    # Reset counters
                    self.processed_requests_count = 0
                    self.last_processing_rate_update = current_time

                # macOS-specific metrics (collected less frequently to reduce overhead)
                if platform.system() == 'Darwin':
                    # Collect GPU metrics
                    gpu_metrics = self._collect_macos_gpu_metrics()
                    if 'gpu_active_residency' in gpu_metrics:
                        self.gpu_active_residency.set(gpu_metrics['gpu_active_residency'])
                    if 'gpu_power' in gpu_metrics:
                        self.gpu_power.set(gpu_metrics['gpu_power'])

                    # Collect power metrics
                    power_metrics = self._collect_macos_power_metrics()
                    if 'cpu_power' in power_metrics:
                        self.cpu_power.set(power_metrics['cpu_power'])
                    if 'package_power' in power_metrics:
                        self.package_power.set(power_metrics['package_power'])

                    # Collect temperature metrics
                    temp_metrics = self._collect_macos_temperature_metrics()
                    if 'thermal_pressure' in temp_metrics:
                        self.thermal_pressure.set(temp_metrics['thermal_pressure'])

            except Exception as e:
                logger.error(f"Error collecting system metrics: {e}")

            await asyncio.sleep(10)

    async def metrics_handler(self, request):
        """Serve Prometheus metrics"""
        metrics = generate_latest()
        return web.Response(
            body=metrics,
            content_type='text/plain; version=0.0.4',
            charset='utf-8'
        )

    def _record_endpoint_latency(self, path, duration, model, body_json):
        """Record latency for specific endpoints"""
        streaming = "true" if (body_json and body_json.get('stream', False)) else "false"
        
        if '/api/generate' in path:
            self.generate_latency.labels(model=model, streaming=streaming).observe(duration)
        elif '/api/chat' in path:
            self.chat_latency.labels(model=model, streaming=streaming).observe(duration)
        elif '/api/tags' in path:
            self.tags_latency.observe(duration)
        elif '/api/show' in path:
            self.show_latency.labels(model=model).observe(duration)

    async def health_handler(self, request):
        """Comprehensive health check endpoint"""
        try:
            health_data = await get_health()
            status_code = 200 if health_data['status'] == 'healthy' else 503
            return web.json_response(health_data, status=status_code)
        except Exception as e:
            logger.error(f"Health check failed: {e}")
            return web.json_response({
                'status': 'unhealthy',
                'error': str(e),
                'proxy_url': f"http://localhost:{self.proxy_port}",
                'metrics_url': f"http://localhost:{self.metrics_port}/metrics",
                'ollama_backend': self.ollama_url
            }, status=503)
    
    async def health_simple_handler(self, request):
        """Simple health check endpoint (fast response)"""
        try:
            health_data = await get_health_simple()
            status_code = 200 if health_data['status'] == 'healthy' else 503
            return web.json_response(health_data, status=status_code)
        except Exception as e:
            logger.error(f"Simple health check failed: {e}")
            return web.json_response({
                'status': 'unhealthy',
                'error': str(e)
            }, status=503)
    
    async def readiness_handler(self, request):
        """Readiness check endpoint (for container orchestration)"""
        try:
            readiness_data = get_readiness()
            status_code = 200 if readiness_data['ready'] else 503
            return web.json_response(readiness_data, status=status_code)
        except Exception as e:
            logger.error(f"Readiness check failed: {e}")
            return web.json_response({
                'ready': False,
                'error': str(e)
            }, status=503)
    
    async def liveness_handler(self, request):
        """Liveness check endpoint (for container orchestration)"""
        try:
            liveness_data = get_liveness()
            status_code = 200 if liveness_data['alive'] else 503
            return web.json_response(liveness_data, status=status_code)
        except Exception as e:
            logger.error(f"Liveness check failed: {e}")
            return web.json_response({
                'alive': False,
                'error': str(e)
            }, status=503)
    
    async def metrics_summary_handler(self, request):
        """Provide metrics summary in JSON format"""
        try:
            from datetime import datetime, timezone
            import statistics
            
            # Get current system metrics
            cpu_percent = psutil.cpu_percent(interval=0.1)
            memory = psutil.virtual_memory()
            
            # Calculate some aggregate metrics
            summary = {
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "summary": {
                    "total_requests": self.processed_requests_count,
                    "active_requests": self.queue_size,
                    "queue_size": self.queue_size,
                    "avg_latency": 0.0,  # Would need to track this
                    "tokens_per_second": 0.0,  # Would need to calculate from recent samples
                    "success_rate": 0.95,  # Placeholder - would calculate from metrics
                    "error_rate": 0.05
                },
                "system": {
                    "cpu_percent": round(cpu_percent, 2),
                    "memory_percent": round(memory.percent, 2),
                    "memory_available_gb": round(memory.available / (1024**3), 2),
                    "memory_used_gb": round(memory.used / (1024**3), 2),
                    "gpu_utilization_percent": 0.0,  # Placeholder
                    "power_cpu_watts": 0.0,  # Placeholder
                    "power_gpu_watts": 0.0   # Placeholder
                },
                "models": {
                    # Would populate with actual model metrics
                }
            }
            
            return web.json_response(summary)
        except Exception as e:
            logger.error(f"Metrics summary failed: {e}")
            return web.json_response({
                'error': str(e)
            }, status=500)

    async def start(self):
        """Start the proxy and metrics servers"""
        # Create proxy app
        proxy_app = web.Application()
        proxy_app.router.add_route('*', '/{path:.*}', self.proxy_request)

        # Create metrics app
        metrics_app = web.Application()
        metrics_app.router.add_get('/metrics', self.metrics_handler)
        metrics_app.router.add_get('/api/metrics/summary', self.metrics_summary_handler)
        
        # Health check endpoints
        metrics_app.router.add_get('/health', self.health_handler)
        metrics_app.router.add_get('/health/simple', self.health_simple_handler)
        metrics_app.router.add_get('/ready', self.readiness_handler)
        metrics_app.router.add_get('/live', self.liveness_handler)

        # Start system metrics collection
        asyncio.create_task(self.collect_system_metrics())

        # Start both servers
        runner1 = web.AppRunner(proxy_app)
        runner2 = web.AppRunner(metrics_app)

        await runner1.setup()
        await runner2.setup()

        site1 = web.TCPSite(runner1, 'localhost', self.proxy_port)
        site2 = web.TCPSite(runner2, 'localhost', self.metrics_port)

        await site1.start()
        await site2.start()

        logger.info(f"🚀 Ollama Monitoring Proxy Started")
        logger.info(f"🔄 Proxy listening on http://localhost:{self.proxy_port}")
        logger.info(f"📊 Metrics available at http://localhost:{self.metrics_port}/metrics")
        logger.info(f"🎯 Forwarding requests to {self.ollama_url}")
        logger.info(f"Use proxy URL in your applications for monitoring")

        # Keep running
        await asyncio.Event().wait()

async def main():
    import argparse
    parser = argparse.ArgumentParser(description='Ollama Monitoring Proxy')
    parser.add_argument('--ollama-host', default='localhost', help='Ollama host')
    parser.add_argument('--ollama-port', type=int, default=11434, help='Ollama port')
    parser.add_argument('--proxy-port', type=int, default=11435, help='Proxy port')
    parser.add_argument('--metrics-port', type=int, default=8001, help='Metrics port')
    parser.add_argument('--portkey-host', default='localhost', help='Portkey host')
    parser.add_argument('--portkey-port', type=int, default=8787, help='Portkey port')
    parser.add_argument('--enable-portkey', action='store_true', help='Route traffic through Portkey')
    
    args = parser.parse_args()
    
    proxy = OllamaMonitoringProxy(
        ollama_host=args.ollama_host,
        ollama_port=args.ollama_port,
        proxy_port=args.proxy_port,
        metrics_port=args.metrics_port,
        portkey_host=args.portkey_host,
        portkey_port=args.portkey_port,
        enable_portkey=args.enable_portkey
    )
    
    if args.enable_portkey:
        logger.info(f"🚪 Portkey routing enabled - forwarding to {proxy.portkey_url}")
    else:
        logger.info(f"🎯 Direct routing - forwarding to {proxy.ollama_url}")
    
    await proxy.start()

if __name__ == '__main__':
    asyncio.run(main())