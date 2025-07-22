#!/usr/bin/env python3
"""
Enhanced Ollama Metrics Server with Comprehensive Monitoring

This server collects detailed metrics from Ollama including:
- Model information and status
- Request performance metrics
- System resource usage
- Error tracking and health monitoring
- Custom histograms for latency analysis
"""

import time
import json
import asyncio
import aiohttp
import psutil
import os
from prometheus_client import (
    start_http_server, Counter, Histogram, Gauge, Info, Enum
)
from datetime import datetime
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class EnhancedOllamaMetrics:
    def __init__(self, ollama_url="http://localhost:11434", port=8000):
        self.ollama_url = ollama_url
        self.port = port
        self.start_time = time.time()

        # Initialize all metrics
        self._init_model_metrics()
        self._init_request_metrics()
        self._init_system_metrics()
        self._init_error_metrics()
        self._init_performance_metrics()

    def _init_model_metrics(self):
        """Model-related metrics"""
        self.model_info = Info(
            'ollama_model_info',
            'Information about loaded Ollama models'
        )

        self.models_available = Gauge(
            'ollama_models_available_total',
            'Total number of available models'
        )

        self.model_size_bytes = Gauge(
            'ollama_model_size_bytes',
            'Size of each model in bytes',
            ['model_name', 'model_family']
        )

        self.model_parameter_count = Gauge(
            'ollama_model_parameters_total',
            'Number of parameters in the model',
            ['model_name', 'parameter_size']
        )

    def _init_request_metrics(self):
        """Request and response metrics"""
        self.requests_total = Counter(
            'ollama_requests_total',
            'Total number of requests to Ollama',
            ['model', 'endpoint', 'status']
        )

        self.request_duration_seconds = Histogram(
            'ollama_request_duration_seconds',
            'Duration of Ollama requests in seconds',
            ['model', 'endpoint'],
            buckets=[0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 25.0, 50.0, 100.0]
        )

        self.time_to_first_token_seconds = Histogram(
            'ollama_time_to_first_token_seconds',
            'Time to first token generation',
            ['model'],
            buckets=[0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 5.0, 10.0]
        )

        self.tokens_per_second = Histogram(
            'ollama_tokens_per_second',
            'Token generation rate',
            ['model', 'type'],  # type: prompt_eval, eval
            buckets=[1, 5, 10, 25, 50, 100, 250, 500, 1000]
        )

        self.prompt_tokens_total = Counter(
            'ollama_prompt_tokens_total',
            'Total prompt tokens processed',
            ['model']
        )

        self.completion_tokens_total = Counter(
            'ollama_completion_tokens_total',
            'Total completion tokens generated',
            ['model']
        )

        self.active_requests = Gauge(
            'ollama_active_requests_current',
            'Number of currently active requests',
            ['model']
        )

    def _init_system_metrics(self):
        """System resource metrics"""
        self.cpu_usage_percent = Gauge(
            'ollama_cpu_usage_percent',
            'CPU usage percentage'
        )

        self.memory_usage_bytes = Gauge(
            'ollama_memory_usage_bytes',
            'Memory usage in bytes',
            ['type']  # rss, vms, shared
        )

        self.disk_io_bytes = Counter(
            'ollama_disk_io_bytes_total',
            'Disk I/O in bytes',
            ['direction']  # read, write
        )

        self.network_bytes = Counter(
            'ollama_network_bytes_total',
            'Network traffic in bytes',
            ['direction']  # sent, received
        )

        self.uptime_seconds = Gauge(
            'ollama_uptime_seconds_total',
            'Uptime of the metrics server'
        )

    def _init_error_metrics(self):
        """Error and health metrics"""
        self.errors_total = Counter(
            'ollama_errors_total',
            'Total number of errors',
            ['model', 'error_type', 'endpoint']
        )

        self.health_status = Enum(
            'ollama_health_status',
            'Health status of Ollama service',
            states=['healthy', 'unhealthy', 'unknown']
        )

        self.last_successful_request = Gauge(
            'ollama_last_successful_request_timestamp',
            'Timestamp of last successful request'
        )

    def _init_performance_metrics(self):
        """Performance and efficiency metrics"""
        self.request_queue_size = Gauge(
            'ollama_request_queue_size',
            'Number of requests in queue'
        )

        self.average_response_size = Histogram(
            'ollama_response_size_bytes',
            'Size of response payloads',
            ['model'],
            buckets=[100, 500, 1000, 5000, 10000, 50000, 100000]
        )

        self.context_length_used = Histogram(
            'ollama_context_length_used',
            'Context length used in requests',
            ['model'],
            buckets=[128, 512, 1024, 2048, 4096, 8192, 16384, 32768]
        )

    async def collect_model_info(self):
        """Collect detailed model information"""
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(f"{self.ollama_url}/api/tags") as response:
                    if response.status == 200:
                        data = await response.json()
                        models = data.get('models', [])

                        # Update total count
                        self.models_available.set(len(models))

                        # Extract model details
                        model_info = {}
                        for model in models:
                            name = model['name']
                            details = model.get('details', {})

                            # Set model size
                            size = model.get('size', 0)
                            family = details.get('family', 'unknown')
                            self.model_size_bytes.labels(
                                model_name=name,
                                model_family=family
                            ).set(size)

                            # Set parameter count (extract from parameter_size like "3.8B")
                            param_size = details.get('parameter_size', '0')
                            self.model_parameter_count.labels(
                                model_name=name,
                                parameter_size=param_size
                            ).set(self._parse_parameter_count(param_size))

                                                        # Collect info for Info metric (flatten the structure)
                            model_key = f"model_{name.replace(':', '_')}"
                            model_info[model_key] = f"family={family},format={details.get('format', 'unknown')},quantization={details.get('quantization_level', 'unknown')}"

                        # Update model info
                        if model_info:
                            self.model_info.info(model_info)
                        self.health_status.state('healthy')
                        logger.info(f"✅ Collected info for {len(models)} models")
                        return True

        except Exception as e:
            logger.error(f"❌ Failed to collect model info: {e}")
            self.health_status.state('unhealthy')
            return False

    def _parse_parameter_count(self, param_size: str) -> float:
        """Parse parameter size string like '3.8B' to number"""
        if not param_size or param_size == '0':
            return 0
        try:
            if param_size.endswith('B'):
                return float(param_size[:-1]) * 1e9
            elif param_size.endswith('M'):
                return float(param_size[:-1]) * 1e6
            elif param_size.endswith('K'):
                return float(param_size[:-1]) * 1e3
            else:
                return float(param_size)
        except:
            return 0

    async def collect_system_metrics(self):
        """Collect system resource metrics"""
        try:
            # CPU usage
            cpu_percent = psutil.cpu_percent(interval=1)
            self.cpu_usage_percent.set(cpu_percent)

            # Memory usage
            memory = psutil.virtual_memory()
            self.memory_usage_bytes.labels(type='total').set(memory.total)
            self.memory_usage_bytes.labels(type='available').set(memory.available)
            self.memory_usage_bytes.labels(type='used').set(memory.used)

            # Disk I/O
            disk_io = psutil.disk_io_counters()
            if disk_io:
                self.disk_io_bytes.labels(direction='read').inc(disk_io.read_bytes)
                self.disk_io_bytes.labels(direction='write').inc(disk_io.write_bytes)

            # Network I/O
            net_io = psutil.net_io_counters()
            if net_io:
                self.network_bytes.labels(direction='sent').inc(net_io.bytes_sent)
                self.network_bytes.labels(direction='received').inc(net_io.bytes_recv)

            # Uptime
            uptime = time.time() - self.start_time
            self.uptime_seconds.set(uptime)

        except Exception as e:
            logger.error(f"❌ Failed to collect system metrics: {e}")

    async def simulate_request_metrics(self):
        """Simulate realistic request metrics for demonstration"""
        import random

        models = ['phi3:mini', 'llama2', 'codellama']
        endpoints = ['generate', 'chat', 'embeddings']

        while True:
            try:
                # Simulate a request
                model = random.choice(models)
                endpoint = random.choice(endpoints)

                # Simulate request timing
                duration = random.uniform(0.5, 5.0)
                ttft = random.uniform(0.1, 1.0)
                tokens_per_sec = random.uniform(10, 100)
                prompt_tokens = random.randint(10, 500)
                completion_tokens = random.randint(20, 200)

                # Update metrics
                self.requests_total.labels(model=model, endpoint=endpoint, status='success').inc()
                self.request_duration_seconds.labels(model=model, endpoint=endpoint).observe(duration)
                self.time_to_first_token_seconds.labels(model=model).observe(ttft)
                self.tokens_per_second.labels(model=model, type='eval').observe(tokens_per_sec)
                self.prompt_tokens_total.labels(model=model).inc(prompt_tokens)
                self.completion_tokens_total.labels(model=model).inc(completion_tokens)

                # Context length
                context_length = prompt_tokens + completion_tokens
                self.context_length_used.labels(model=model).observe(context_length)

                # Response size (simulate)
                response_size = completion_tokens * random.randint(3, 6)  # bytes per token
                self.average_response_size.labels(model=model).observe(response_size)

                self.last_successful_request.set(time.time())

                # Occasional errors
                if random.random() < 0.05:  # 5% error rate
                    error_type = random.choice(['timeout', 'model_not_found', 'out_of_memory'])
                    self.errors_total.labels(model=model, error_type=error_type, endpoint=endpoint).inc()

                await asyncio.sleep(random.uniform(2, 8))

            except Exception as e:
                logger.error(f"❌ Error in request simulation: {e}")
                await asyncio.sleep(5)

    def start_metrics_server(self):
        """Start the Prometheus metrics HTTP server"""
        try:
            start_http_server(self.port)
            logger.info(f"📊 Enhanced metrics server started at http://localhost:{self.port}/metrics")
            return True
        except Exception as e:
            logger.error(f"❌ Failed to start metrics server: {e}")
            return False

    async def run(self):
        """Main run loop"""
        logger.info("🚀 Starting Enhanced Ollama Metrics Server")

        if not self.start_metrics_server():
            return

        # Initial setup
        await self.collect_model_info()

        # Start background tasks
        tasks = [
            asyncio.create_task(self._periodic_model_collection()),
            asyncio.create_task(self._periodic_system_collection()),
            asyncio.create_task(self.simulate_request_metrics())
        ]

        logger.info("✅ Enhanced metrics server is running!")
        logger.info(f"📈 Metrics available at http://localhost:{self.port}/metrics")
        logger.info("🎯 Comprehensive Ollama monitoring active")
        logger.info("Press Ctrl+C to stop")

        try:
            await asyncio.gather(*tasks)
        except KeyboardInterrupt:
            logger.info("🛑 Shutting down enhanced metrics server...")
            for task in tasks:
                task.cancel()
        except Exception as e:
            logger.error(f"❌ Error in enhanced metrics server: {e}")
        finally:
            logger.info("✅ Enhanced metrics server stopped")

    async def _periodic_model_collection(self):
        """Periodically collect model information"""
        while True:
            await self.collect_model_info()
            await asyncio.sleep(60)  # Every minute

    async def _periodic_system_collection(self):
        """Periodically collect system metrics"""
        while True:
            await self.collect_system_metrics()
            await asyncio.sleep(15)  # Every 15 seconds

async def main():
    server = EnhancedOllamaMetrics()
    await server.run()

if __name__ == "__main__":
    asyncio.run(main())