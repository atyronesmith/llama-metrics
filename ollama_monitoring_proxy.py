#!/usr/bin/env python3
"""
Ollama Monitoring Proxy Server
Implements comprehensive monitoring for Ollama with Prometheus metrics
"""

import time
import json
import asyncio
import aiohttp
from aiohttp import web
import logging
from datetime import datetime
from prometheus_client import (
    Counter, Histogram, Gauge, Info, generate_latest
)
import psutil
from typing import Dict, Any, Optional

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class OllamaMonitoringProxy:
    def __init__(self, ollama_url="http://localhost:11434", proxy_port=11435, metrics_port=8001):
        self.ollama_url = ollama_url
        self.proxy_port = proxy_port
        self.metrics_port = metrics_port

        # Initialize metrics
        self._init_metrics()

    def _init_metrics(self):
        """Initialize all Prometheus metrics"""

        # Request metrics
        self.request_count = Counter(
            'ollama_proxy_requests_total',
            'Total number of requests through proxy',
            ['method', 'endpoint', 'model', 'status']
        )

        self.request_duration = Histogram(
            'ollama_proxy_request_duration_seconds',
            'Request duration in seconds',
            ['method', 'endpoint', 'model'],
            buckets=[0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 25.0, 50.0, 100.0, 250.0]
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
            'Token generation rate',
            ['model', 'phase'],  # phase: prompt_eval, eval
            buckets=[1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500]
        )

        # Performance metrics
        self.time_to_first_token = Histogram(
            'ollama_proxy_time_to_first_token_seconds',
            'Time to first token',
            ['model'],
            buckets=[0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 5.0, 10.0]
        )

        self.model_load_duration = Histogram(
            'ollama_proxy_model_load_duration_seconds',
            'Model loading duration',
            ['model'],
            buckets=[0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0, 120.0]
        )

        # Error tracking
        self.errors = Counter(
            'ollama_proxy_errors_total',
            'Total errors',
            ['model', 'error_type', 'endpoint']
        )

        # System metrics
        self.cpu_usage = Gauge(
            'ollama_proxy_cpu_usage_percent',
            'CPU usage percentage'
        )

        self.memory_usage = Gauge(
            'ollama_proxy_memory_usage_bytes',
            'Memory usage in bytes',
            ['type']  # rss, vms, percent
        )

        # Queue metrics
        self.queue_size = Gauge(
            'ollama_proxy_queue_size',
            'Number of requests in queue'
        )

        # Context length metrics
        self.context_length = Histogram(
            'ollama_proxy_context_length',
            'Context length used',
            ['model'],
            buckets=[128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768]
        )

    async def proxy_request(self, request):
        """Proxy requests to Ollama and collect metrics"""
        path = request.path_qs
        method = request.method
        model = "unknown"
        start_time = time.time()

        # Extract model from request body if present
        if request.can_read_body:
            try:
                body = await request.read()
                body_json = json.loads(body) if body else {}
                model = body_json.get('model', 'unknown')
            except:
                body = await request.read()
                body_json = {}
        else:
            body = None
            body_json = {}

        # Update active requests
        self.active_requests.labels(model=model).inc()

        try:
            # Prepare headers - filter out hop-by-hop headers and content-type
            # We'll let aiohttp handle content-type when using json parameter
            hop_by_hop_headers = {
                'host', 'content-length', 'transfer-encoding',
                'connection', 'keep-alive', 'proxy-authenticate',
                'proxy-authorization', 'te', 'trailers', 'upgrade',
                'content-type'  # Let aiohttp set this when using json param
            }
            headers = {k: v for k, v in request.headers.items()
                      if k.lower() not in hop_by_hop_headers}

            # Preserve content-type only for non-JSON requests
            if not body_json and 'content-type' in request.headers:
                headers['content-type'] = request.headers['content-type']

            # Make request to Ollama
            async with aiohttp.ClientSession() as session:
                url = f"{self.ollama_url}{path}"

                # Use json parameter for JSON payloads to avoid header conflicts
                kwargs = {
                    'method': method,
                    'url': url,
                    'headers': headers,
                    'timeout': aiohttp.ClientTimeout(total=300)
                }

                # If we have JSON data, use json parameter instead of data
                if body and body_json:
                    kwargs['json'] = body_json
                elif body:
                    kwargs['data'] = body

                async with session.request(**kwargs) as resp:

                    # Handle streaming responses
                    if 'stream' in body_json and body_json['stream']:
                        return await self._handle_streaming_response(
                            resp, request, model, start_time, path
                        )
                    else:
                        # Non-streaming response
                        response_body = await resp.read()
                        response_json = json.loads(response_body) if response_body else {}

                        # Extract metrics from response
                        self._extract_response_metrics(response_json, model, start_time)

                        # Update request metrics
                        duration = time.time() - start_time
                        self.request_duration.labels(
                            method=method, endpoint=path, model=model
                        ).observe(duration)
                        self.request_count.labels(
                            method=method, endpoint=path, model=model, status=resp.status
                        ).inc()

                        return web.Response(
                            body=response_body,
                            status=resp.status,
                            headers=resp.headers
                        )

        except asyncio.TimeoutError:
            self.errors.labels(model=model, error_type='timeout', endpoint=path).inc()
            return web.Response(text="Request timeout", status=504)
        except Exception as e:
            self.errors.labels(model=model, error_type=type(e).__name__, endpoint=path).inc()
            logger.error(f"Proxy error: {e}")
            return web.Response(text=str(e), status=500)
        finally:
            self.active_requests.labels(model=model).dec()

    async def _handle_streaming_response(self, resp, request, model, start_time, path):
        """Handle streaming responses with metrics collection"""
        response = web.StreamResponse(
            status=resp.status,
            headers=resp.headers
        )
        await response.prepare(request)

        first_token_time = None
        total_tokens = 0
        prompt_eval_count = 0

        try:
            async for chunk in resp.content.iter_any():
                await response.write(chunk)

                # Try to parse streaming chunks for metrics
                try:
                    lines = chunk.decode('utf-8').strip().split('\n')
                    for line in lines:
                        if line:
                            data = json.loads(line)

                            # Time to first token
                            if first_token_time is None and data.get('response'):
                                first_token_time = time.time()
                                ttft = first_token_time - start_time
                                self.time_to_first_token.labels(model=model).observe(ttft)

                            # Token counts
                            if 'eval_count' in data:
                                total_tokens = data['eval_count']
                            if 'prompt_eval_count' in data:
                                prompt_eval_count = data['prompt_eval_count']

                            # Final metrics
                            if data.get('done'):
                                self._extract_response_metrics(data, model, start_time)

                except:
                    pass  # Ignore parsing errors for chunks

        finally:
            await response.write_eof()

            # Update request metrics
            duration = time.time() - start_time
            self.request_duration.labels(
                method='POST', endpoint=path, model=model
            ).observe(duration)
            self.request_count.labels(
                method='POST', endpoint=path, model=model, status=resp.status
            ).inc()

        return response

    def _extract_response_metrics(self, response_json: Dict[str, Any], model: str, start_time: float):
        """Extract metrics from Ollama response"""
        # Token metrics
        if 'prompt_eval_count' in response_json:
            self.prompt_tokens.labels(model=model).inc(response_json['prompt_eval_count'])

        if 'eval_count' in response_json:
            self.generated_tokens.labels(model=model).inc(response_json['eval_count'])

        # Token generation rates
        if 'prompt_eval_duration' in response_json and response_json.get('prompt_eval_count'):
            prompt_eval_duration = response_json['prompt_eval_duration'] / 1e9  # Convert nanoseconds
            if prompt_eval_duration > 0:
                prompt_tokens_per_sec = response_json['prompt_eval_count'] / prompt_eval_duration
                self.tokens_per_second.labels(model=model, phase='prompt_eval').observe(prompt_tokens_per_sec)

        if 'eval_duration' in response_json and response_json.get('eval_count'):
            eval_duration = response_json['eval_duration'] / 1e9  # Convert nanoseconds
            if eval_duration > 0:
                eval_tokens_per_sec = response_json['eval_count'] / eval_duration
                self.tokens_per_second.labels(model=model, phase='eval').observe(eval_tokens_per_sec)

        # Model load time
        if 'load_duration' in response_json:
            load_duration = response_json['load_duration'] / 1e9
            self.model_load_duration.labels(model=model).observe(load_duration)

        # Context length
        total_context = response_json.get('prompt_eval_count', 0) + response_json.get('eval_count', 0)
        if total_context > 0:
            self.context_length.labels(model=model).observe(total_context)

    async def collect_system_metrics(self):
        """Continuously collect system metrics"""
        while True:
            try:
                # CPU usage
                cpu_percent = psutil.cpu_percent(interval=1)
                self.cpu_usage.set(cpu_percent)

                # Memory usage
                memory = psutil.virtual_memory()
                process = psutil.Process()
                memory_info = process.memory_info()

                self.memory_usage.labels(type='rss').set(memory_info.rss)
                self.memory_usage.labels(type='vms').set(memory_info.vms)
                self.memory_usage.labels(type='percent').set(memory.percent)

            except Exception as e:
                logger.error(f"Error collecting system metrics: {e}")

            await asyncio.sleep(10)  # Collect every 10 seconds

    async def metrics_handler(self, request):
        """Prometheus metrics endpoint handler"""
        metrics = generate_latest()
        return web.Response(
            body=metrics,
            content_type='text/plain; version=0.0.4',
            charset='utf-8'
        )

    async def health_handler(self, request):
        """Health check endpoint"""
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(f"{self.ollama_url}/api/tags", timeout=aiohttp.ClientTimeout(total=5)) as resp:
                    if resp.status == 200:
                        return web.json_response({"status": "healthy", "ollama": "connected"})
        except:
            pass

        return web.json_response({"status": "unhealthy", "ollama": "disconnected"}, status=503)

    def create_app(self):
        """Create aiohttp application"""
        app = web.Application()

        # Add routes
        app.router.add_route('*', '/metrics', self.metrics_handler)
        app.router.add_route('GET', '/health', self.health_handler)
        app.router.add_route('*', '/{path:.*}', self.proxy_request)

        return app

    async def start(self):
        """Start proxy and metrics servers"""
        # Start system metrics collection
        asyncio.create_task(self.collect_system_metrics())

        # Create apps
        proxy_app = self.create_app()
        metrics_app = web.Application()
        metrics_app.router.add_get('/metrics', self.metrics_handler)

        # Start servers
        runner1 = web.AppRunner(proxy_app)
        runner2 = web.AppRunner(metrics_app)

        await runner1.setup()
        await runner2.setup()

        site1 = web.TCPSite(runner1, '0.0.0.0', self.proxy_port)
        site2 = web.TCPSite(runner2, '0.0.0.0', self.metrics_port)

        await site1.start()
        await site2.start()

        logger.info("🚀 Ollama Monitoring Proxy Started")
        logger.info(f"🔄 Proxy listening on http://localhost:{self.proxy_port}")
        logger.info(f"📊 Metrics available at http://localhost:{self.metrics_port}/metrics")
        logger.info(f"🎯 Forwarding requests to {self.ollama_url}")
        logger.info("Use proxy URL in your applications for monitoring")

        # Keep running
        await asyncio.Event().wait()

async def main():
    proxy = OllamaMonitoringProxy()
    await proxy.start()

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("🛑 Shutting down proxy...")