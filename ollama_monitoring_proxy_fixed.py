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
from aiohttp import web
from prometheus_client import (
    Counter, Histogram, Gauge, Info, generate_latest
)

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class OllamaMonitoringProxy:
    def __init__(self, ollama_host='localhost', ollama_port=11434,
                 proxy_port=11435, metrics_port=8001):
        self.ollama_url = f"http://{ollama_host}:{ollama_port}"
        self.proxy_port = proxy_port
        self.metrics_port = metrics_port

        # Initialize metrics
        self._init_metrics()

        # Track queue size
        self.queue_size = 0

    def _init_metrics(self):
        """Initialize Prometheus metrics"""
        # Request metrics
        self.request_count = Counter(
            'ollama_proxy_requests_total',
            'Total number of requests',
            ['method', 'endpoint', 'model', 'status']
        )

        self.request_duration = Histogram(
            'ollama_proxy_request_duration_seconds',
            'Request duration in seconds',
            ['method', 'endpoint', 'model'],
            buckets=(0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0, 120.0)
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

        # Update active requests
        self.active_requests.labels(model=model).inc()

        try:
            # Create a new session for the proxy request
            async with aiohttp.ClientSession() as session:
                url = f"{self.ollama_url}{path}"

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
                        self.request_duration.labels(
                            method=method, endpoint=path, model=model
                        ).observe(duration)

                        self.request_count.labels(
                            method=method, endpoint=path, model=model,
                            status=resp.status
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
            self.active_requests.labels(model=model).dec()

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
        self.request_duration.labels(
            method='POST', endpoint=path, model=model
        ).observe(duration)

        self.request_count.labels(
            method='POST', endpoint=path, model=model, status=resp.status
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

    async def health_handler(self, request):
        """Health check endpoint"""
        return web.json_response({
            'status': 'healthy',
            'proxy_url': f"http://localhost:{self.proxy_port}",
            'metrics_url': f"http://localhost:{self.metrics_port}/metrics",
            'ollama_backend': self.ollama_url
        })

    async def start(self):
        """Start the proxy and metrics servers"""
        # Create proxy app
        proxy_app = web.Application()
        proxy_app.router.add_route('*', '/{path:.*}', self.proxy_request)

        # Create metrics app
        metrics_app = web.Application()
        metrics_app.router.add_get('/metrics', self.metrics_handler)
        metrics_app.router.add_get('/health', self.health_handler)

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
    proxy = OllamaMonitoringProxy()
    await proxy.start()

if __name__ == '__main__':
    asyncio.run(main())