#!/usr/bin/env python3
"""
Simple Prometheus Metrics Server for Ollama Monitoring

This lightweight server exposes metrics that can be scraped by Prometheus
to monitor AI model usage, without requiring heavy LlamaIndex dependencies.
"""

import time
import json
import asyncio
import aiohttp
from prometheus_client import start_http_server, Counter, Histogram, Gauge, Info
from datetime import datetime
import logging

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Prometheus Metrics
OLLAMA_REQUESTS_TOTAL = Counter(
    'ollama_requests_total',
    'Total number of requests to Ollama',
    ['model', 'status']
)

OLLAMA_REQUEST_DURATION = Histogram(
    'ollama_request_duration_seconds',
    'Duration of Ollama requests in seconds',
    ['model']
)

OLLAMA_ACTIVE_REQUESTS = Gauge(
    'ollama_active_requests',
    'Number of active Ollama requests',
    ['model']
)

OLLAMA_MODEL_INFO = Info(
    'ollama_model_info',
    'Information about available Ollama models'
)

# Metrics for traffic generator monitoring
TRAFFIC_QUESTIONS_ASKED = Counter(
    'traffic_questions_asked_total',
    'Total questions asked by traffic generator',
    ['category']
)

class SimpleMetricsServer:
    def __init__(self, ollama_url="http://localhost:11434", port=8000):
        self.ollama_url = ollama_url
        self.port = port
        self.metrics_data = {
            "requests_total": 0,
            "requests_successful": 0,
            "requests_failed": 0,
            "last_request_time": None,
            "models_available": []
        }

    async def check_ollama_health(self):
        """Check if Ollama is accessible and update model info"""
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(f"{self.ollama_url}/api/tags") as response:
                    if response.status == 200:
                        data = await response.json()
                        models = [model['name'] for model in data.get('models', [])]
                        self.metrics_data["models_available"] = models

                        # Update model info metric
                        OLLAMA_MODEL_INFO.info({
                            'available_models': ','.join(models),
                            'total_models': str(len(models)),
                            'last_check': datetime.now().isoformat()
                        })

                        logger.info(f"✅ Ollama healthy. Models: {', '.join(models)}")
                        return True
                    else:
                        logger.warning(f"⚠️ Ollama returned status {response.status}")
                        return False
        except Exception as e:
            logger.error(f"❌ Cannot connect to Ollama: {e}")
            return False

    async def simulate_traffic_metrics(self):
        """Simulate some traffic metrics for demonstration"""
        categories = ["general_knowledge", "science", "technology", "history", "geography"]

        while True:
            # Simulate random question activity
            import random
            category = random.choice(categories)
            TRAFFIC_QUESTIONS_ASKED.labels(category=category).inc()

            # Simulate an Ollama request
            model = "phi3:mini"
            start_time = time.time()

            # Simulate request duration (between 0.5 and 3 seconds)
            duration = random.uniform(0.5, 3.0)
            await asyncio.sleep(0.1)  # Brief pause

            # Record metrics
            OLLAMA_REQUESTS_TOTAL.labels(model=model, status="success").inc()
            OLLAMA_REQUEST_DURATION.labels(model=model).observe(duration)

            self.metrics_data["requests_total"] += 1
            self.metrics_data["requests_successful"] += 1
            self.metrics_data["last_request_time"] = datetime.now().isoformat()

            # Wait before next simulated request (5-15 seconds)
            await asyncio.sleep(random.uniform(5, 15))

    async def health_check_loop(self):
        """Periodically check Ollama health"""
        while True:
            await self.check_ollama_health()
            await asyncio.sleep(30)  # Check every 30 seconds

    def start_metrics_server(self):
        """Start the Prometheus metrics HTTP server"""
        try:
            start_http_server(self.port)
            logger.info(f"📊 Prometheus metrics server started at http://localhost:{self.port}/metrics")
            return True
        except Exception as e:
            logger.error(f"❌ Failed to start metrics server: {e}")
            return False

    async def run(self):
        """Main run loop"""
        logger.info("🚀 Starting Simple Ollama Metrics Server")

        # Start Prometheus metrics server
        if not self.start_metrics_server():
            return

        # Initial health check
        await self.check_ollama_health()

        # Start background tasks
        health_task = asyncio.create_task(self.health_check_loop())
        traffic_task = asyncio.create_task(self.simulate_traffic_metrics())

        logger.info("✅ Metrics server is running!")
        logger.info(f"📈 Visit http://localhost:{self.port}/metrics to see metrics")
        logger.info("🎯 Prometheus should scrape this endpoint")
        logger.info("Press Ctrl+C to stop")

        try:
            # Keep running until interrupted
            await asyncio.gather(health_task, traffic_task)
        except KeyboardInterrupt:
            logger.info("🛑 Shutting down metrics server...")
            health_task.cancel()
            traffic_task.cancel()
        except Exception as e:
            logger.error(f"❌ Error in metrics server: {e}")
        finally:
            logger.info("✅ Metrics server stopped")

async def main():
    server = SimpleMetricsServer()
    await server.run()

if __name__ == "__main__":
    asyncio.run(main())