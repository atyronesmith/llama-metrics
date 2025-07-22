#!/usr/bin/env python3
"""
High-Performance Ollama Load Tester
===================================

A comprehensive load testing tool designed to stress test Ollama LLM servers
with configurable concurrency, request rates, and load patterns.

Features:
- Massive concurrent requests (10-1000+ simultaneous)
- Multiple load patterns (constant, burst, ramp-up, spike)
- Real-time statistics and monitoring
- Queue saturation testing
- Customizable request profiles
"""

import asyncio
import aiohttp
import time
import json
import random
import logging
import argparse
import signal
import sys
from datetime import datetime, timedelta
from typing import List, Dict, Optional, Callable, Any
from dataclasses import dataclass, field
from collections import defaultdict, deque
import statistics

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('load_test.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

@dataclass
class LoadTestConfig:
    """Configuration for load testing scenarios"""
    base_url: str = "http://localhost:11435"  # Use monitoring proxy
    model: str = "phi3:mini"
    
    # Concurrency settings
    max_concurrent_requests: int = 5  # Safe default to prevent Ollama overload
    total_requests: int = 100  # Reduced default request count
    duration_seconds: Optional[int] = None
    
    # Request rate settings
    requests_per_second: float = 2.0  # Safer RPS default
    burst_size: int = 20
    burst_interval: float = 5.0
    
    # Load pattern
    pattern: str = "constant"  # constant, burst, ramp, spike, chaos
    
    # Request settings
    stream: bool = False
    timeout: float = 300.0
    
    # Test prompts
    prompt_types: List[str] = field(default_factory=lambda: ["short", "medium", "long"])

@dataclass
class RequestStats:
    """Statistics for individual requests"""
    start_time: float
    end_time: Optional[float] = None
    success: bool = False
    status_code: Optional[int] = None
    error: Optional[str] = None
    response_size: int = 0
    tokens_generated: int = 0
    queue_wait_time: float = 0.0

class LoadTestStats:
    """Real-time statistics tracking"""
    
    def __init__(self):
        self.requests_completed = 0
        self.requests_failed = 0
        self.requests_timeout = 0
        self.total_response_time = 0.0
        self.response_times = deque(maxlen=1000)  # Keep last 1000 for percentiles
        self.start_time = time.time()
        self.last_report_time = time.time()
        self.requests_per_second_history = []
        self.concurrent_requests = 0
        self.max_concurrent = 0
        self.status_codes = defaultdict(int)
        self.errors = defaultdict(int)
        
    def record_request_start(self):
        """Record when a request starts"""
        self.concurrent_requests += 1
        self.max_concurrent = max(self.max_concurrent, self.concurrent_requests)
    
    def record_request_end(self, request_stats: RequestStats):
        """Record completed request statistics"""
        self.concurrent_requests -= 1
        
        if request_stats.success:
            self.requests_completed += 1
            response_time = request_stats.end_time - request_stats.start_time
            self.total_response_time += response_time
            self.response_times.append(response_time)
            self.status_codes[request_stats.status_code] += 1
        else:
            self.requests_failed += 1
            if "timeout" in str(request_stats.error).lower():
                self.requests_timeout += 1
            self.errors[request_stats.error] += 1
    
    def get_current_rps(self) -> float:
        """Calculate current requests per second"""
        current_time = time.time()
        elapsed = current_time - self.last_report_time
        if elapsed > 0:
            recent_requests = len([t for t in self.response_times if current_time - t < elapsed])
            return recent_requests / elapsed
        return 0.0
    
    def get_stats_summary(self) -> Dict[str, Any]:
        """Get comprehensive statistics summary"""
        current_time = time.time()
        total_elapsed = current_time - self.start_time
        total_requests = self.requests_completed + self.requests_failed
        
        avg_response_time = (self.total_response_time / max(1, self.requests_completed))
        success_rate = (self.requests_completed / max(1, total_requests)) * 100
        
        # Calculate percentiles
        percentiles = {}
        if self.response_times:
            sorted_times = sorted(self.response_times)
            percentiles = {
                'p50': statistics.median(sorted_times),
                'p75': sorted_times[int(len(sorted_times) * 0.75)],
                'p90': sorted_times[int(len(sorted_times) * 0.90)],
                'p95': sorted_times[int(len(sorted_times) * 0.95)],
                'p99': sorted_times[int(len(sorted_times) * 0.99)],
                'min': min(sorted_times),
                'max': max(sorted_times)
            }
        
        return {
            'duration': total_elapsed,
            'total_requests': total_requests,
            'completed': self.requests_completed,
            'failed': self.requests_failed,
            'timeouts': self.requests_timeout,
            'success_rate': success_rate,
            'avg_response_time': avg_response_time,
            'current_rps': self.get_current_rps(),
            'avg_rps': total_requests / max(1, total_elapsed),
            'concurrent_requests': self.concurrent_requests,
            'max_concurrent': self.max_concurrent,
            'percentiles': percentiles,
            'status_codes': dict(self.status_codes),
            'top_errors': dict(list(self.errors.items())[:5])
        }

class HighPerformanceLoadTester:
    """High-performance concurrent load tester"""
    
    def __init__(self, config: LoadTestConfig):
        self.config = config
        self.stats = LoadTestStats()
        self.session = None
        self.running = False
        self.prompts = self._generate_test_prompts()
        
        # Setup signal handlers
        signal.signal(signal.SIGINT, self._signal_handler)
        signal.signal(signal.SIGTERM, self._signal_handler)
    
    def _signal_handler(self, signum, frame):
        """Handle shutdown signals gracefully"""
        logger.info("🛑 Received shutdown signal. Stopping load test...")
        self.running = False
        # Force exit after a timeout to avoid hanging
        import threading
        def force_exit():
            import time
            time.sleep(3)  # Give 3 seconds for graceful shutdown
            logger.warning("⚠️  Force exiting due to timeout")
            import os
            os._exit(1)
        threading.Thread(target=force_exit, daemon=True).start()
    
    def _generate_test_prompts(self) -> Dict[str, List[str]]:
        """Generate different types of test prompts"""
        return {
            'short': [
                "What is 2+2?",
                "Name a color.",
                "What day is today?",
                "Count to 3.",
                "Say hello.",
                "What is AI?",
                "Name an animal.",
                "What is Python?",
                "Define JSON.",
                "What is REST?"
            ],
            'medium': [
                "Explain the concept of machine learning in simple terms.",
                "What are the main differences between Python and JavaScript?",
                "How does HTTP work and what are the main HTTP methods?",
                "Describe the process of photosynthesis step by step.",
                "What are the key principles of good software design?",
                "Explain what containerization is and its benefits.",
                "How do databases work and what is SQL?",
                "What is cloud computing and its main service models?",
                "Describe the differences between synchronous and asynchronous programming.",
                "What are microservices and how do they differ from monoliths?"
            ],
            'long': [
                "Write a detailed explanation of how modern web applications work, including the frontend, backend, database layers, and how they communicate. Include information about HTTP protocols, REST APIs, authentication, and security considerations.",
                "Explain the entire machine learning pipeline from data collection to model deployment in production. Include details about data preprocessing, feature engineering, model selection, training, evaluation, and monitoring.",
                "Describe the architecture of modern distributed systems. Explain concepts like load balancing, service discovery, circuit breakers, eventual consistency, CAP theorem, and how to handle failures in distributed environments.",
                "Provide a comprehensive guide to building secure applications. Cover topics like authentication, authorization, input validation, SQL injection prevention, XSS protection, HTTPS, and other security best practices.",
                "Explain how modern operating systems work, including process management, memory management, file systems, I/O operations, scheduling algorithms, and how applications interact with the kernel."
            ]
        }
    
    async def _create_session(self):
        """Create aiohttp session with optimized settings"""
        connector = aiohttp.TCPConnector(
            limit=self.config.max_concurrent_requests * 2,
            limit_per_host=self.config.max_concurrent_requests * 2,
            keepalive_timeout=30,
            enable_cleanup_closed=True
        )
        
        timeout = aiohttp.ClientTimeout(total=self.config.timeout)
        self.session = aiohttp.ClientSession(
            connector=connector,
            timeout=timeout,
            headers={'User-Agent': 'HighPerformanceLoadTester/1.0'}
        )
    
    async def _send_request(self, prompt: str) -> RequestStats:
        """Send a single request and return statistics"""
        request_stats = RequestStats(start_time=time.time())
        
        payload = {
            "model": self.config.model,
            "prompt": prompt,
            "stream": self.config.stream
        }
        
        try:
            async with self.session.post(
                f"{self.config.base_url}/api/generate",
                json=payload
            ) as response:
                request_stats.status_code = response.status
                response_data = await response.read()
                request_stats.response_size = len(response_data)
                request_stats.end_time = time.time()
                
                if response.status == 200:
                    request_stats.success = True
                    # Try to extract token count from response
                    try:
                        data = json.loads(response_data.decode('utf-8'))
                        if 'response' in data:
                            # Rough token estimation: ~4 chars per token
                            request_stats.tokens_generated = len(data['response']) // 4
                    except:
                        pass
                else:
                    request_stats.error = f"HTTP {response.status}"
                    
        except asyncio.TimeoutError:
            request_stats.end_time = time.time()
            request_stats.error = "Request timeout"
        except Exception as e:
            request_stats.end_time = time.time()
            request_stats.error = str(e)
        
        return request_stats
    
    async def _request_worker(self, semaphore: asyncio.Semaphore, request_queue: asyncio.Queue):
        """Worker coroutine that processes requests from queue"""
        while self.running:
            try:
                # Get next request from queue
                prompt = await asyncio.wait_for(request_queue.get(), timeout=1.0)
                
                async with semaphore:
                    self.stats.record_request_start()
                    request_stats = await self._send_request(prompt)
                    self.stats.record_request_end(request_stats)
                    request_queue.task_done()
                    
            except asyncio.TimeoutError:
                # No requests in queue, continue
                continue
            except Exception as e:
                logger.error(f"Request worker error: {e}")
    
    async def _generate_constant_load(self, request_queue: asyncio.Queue):
        """Generate constant load pattern"""
        request_interval = 1.0 / self.config.requests_per_second
        requests_sent = 0
        
        while self.running and requests_sent < self.config.total_requests:
            prompt_type = random.choice(self.config.prompt_types)
            prompt = random.choice(self.prompts[prompt_type])
            
            await request_queue.put(prompt)
            requests_sent += 1
            
            await asyncio.sleep(request_interval)
    
    async def _generate_burst_load(self, request_queue: asyncio.Queue):
        """Generate burst load pattern"""
        requests_sent = 0
        
        while self.running and requests_sent < self.config.total_requests:
            # Send burst of requests
            for _ in range(min(self.config.burst_size, self.config.total_requests - requests_sent)):
                prompt_type = random.choice(self.config.prompt_types)
                prompt = random.choice(self.prompts[prompt_type])
                await request_queue.put(prompt)
                requests_sent += 1
            
            # Wait between bursts
            await asyncio.sleep(self.config.burst_interval)
    
    async def _generate_ramp_load(self, request_queue: asyncio.Queue):
        """Generate ramping load pattern"""
        duration = self.config.duration_seconds or 300
        max_rps = self.config.requests_per_second
        start_time = time.time()
        requests_sent = 0
        
        while self.running and requests_sent < self.config.total_requests:
            # Calculate current RPS based on ramp progress
            elapsed = time.time() - start_time
            progress = min(1.0, elapsed / duration)
            current_rps = max_rps * progress
            
            if current_rps > 0:
                request_interval = 1.0 / current_rps
                
                prompt_type = random.choice(self.config.prompt_types)
                prompt = random.choice(self.prompts[prompt_type])
                await request_queue.put(prompt)
                requests_sent += 1
                
                await asyncio.sleep(request_interval)
            else:
                await asyncio.sleep(0.1)
    
    async def _generate_spike_load(self, request_queue: asyncio.Queue):
        """Generate spike load pattern with sudden bursts"""
        requests_sent = 0
        base_interval = 1.0 / (self.config.requests_per_second * 0.3)  # 30% of target RPS normally
        
        while self.running and requests_sent < self.config.total_requests:
            # Normal load
            for _ in range(10):
                if requests_sent >= self.config.total_requests:
                    break
                prompt_type = random.choice(self.config.prompt_types)
                prompt = random.choice(self.prompts[prompt_type])
                await request_queue.put(prompt)
                requests_sent += 1
                await asyncio.sleep(base_interval)
            
            # Spike - send many requests quickly
            spike_size = min(self.config.burst_size * 3, self.config.total_requests - requests_sent)
            for _ in range(spike_size):
                prompt_type = random.choice(self.config.prompt_types)
                prompt = random.choice(self.prompts[prompt_type])
                await request_queue.put(prompt)
                requests_sent += 1
                await asyncio.sleep(0.01)  # Very fast during spike
    
    async def _generate_chaos_load(self, request_queue: asyncio.Queue):
        """Generate chaotic load pattern with random intervals"""
        requests_sent = 0
        base_interval = 1.0 / self.config.requests_per_second
        
        while self.running and requests_sent < self.config.total_requests:
            # Random interval between 0.1x and 5x the base interval
            interval = base_interval * random.uniform(0.1, 5.0)
            
            # Random burst size
            burst_size = random.randint(1, self.config.burst_size)
            
            for _ in range(min(burst_size, self.config.total_requests - requests_sent)):
                prompt_type = random.choice(self.config.prompt_types)
                prompt = random.choice(self.prompts[prompt_type])
                await request_queue.put(prompt)
                requests_sent += 1
                await asyncio.sleep(random.uniform(0, 0.1))  # Small random delay within burst
            
            await asyncio.sleep(interval)
    
    async def _stats_reporter(self):
        """Report statistics periodically"""
        while self.running:
            await asyncio.sleep(5)  # Report every 5 seconds
            stats = self.stats.get_stats_summary()
            
            logger.info("📊 LOAD TEST STATS:")
            logger.info(f"   Duration: {stats['duration']:.1f}s")
            logger.info(f"   Requests: {stats['completed']}/{stats['total_requests']} "
                       f"({stats['success_rate']:.1f}% success)")
            logger.info(f"   RPS: {stats['current_rps']:.1f} current, {stats['avg_rps']:.1f} average")
            logger.info(f"   Concurrency: {stats['concurrent_requests']} current, "
                       f"{stats['max_concurrent']} peak")
            logger.info(f"   Response Time: {stats['avg_response_time']:.2f}s avg")
            
            if stats['percentiles']:
                p = stats['percentiles']
                logger.info(f"   Latency: p50={p['p50']:.2f}s p95={p['p95']:.2f}s "
                           f"p99={p['p99']:.2f}s max={p['max']:.2f}s")
            
            if stats['failed'] > 0:
                logger.info(f"   Failures: {stats['failed']} ({stats['timeouts']} timeouts)")
    
    async def run_load_test(self):
        """Execute the load test with specified pattern"""
        logger.info("🚀 Starting High-Performance Load Test")
        logger.info(f"   Pattern: {self.config.pattern}")
        logger.info(f"   Target RPS: {self.config.requests_per_second}")
        logger.info(f"   Max Concurrent: {self.config.max_concurrent_requests}")
        logger.info(f"   Total Requests: {self.config.total_requests}")
        logger.info(f"   Proxy URL: {self.config.base_url}")
        
        self.running = True
        
        # Create HTTP session
        await self._create_session()
        
        # Create semaphore to limit concurrency
        semaphore = asyncio.Semaphore(self.config.max_concurrent_requests)
        
        # Create request queue
        request_queue = asyncio.Queue(maxsize=self.config.max_concurrent_requests * 2)
        
        # Start worker tasks
        workers = [
            asyncio.create_task(self._request_worker(semaphore, request_queue))
            for _ in range(self.config.max_concurrent_requests)
        ]
        
        # Start stats reporter
        reporter_task = asyncio.create_task(self._stats_reporter())
        
        # Start load generation based on pattern
        load_generators = {
            'constant': self._generate_constant_load,
            'burst': self._generate_burst_load,
            'ramp': self._generate_ramp_load,
            'spike': self._generate_spike_load,
            'chaos': self._generate_chaos_load
        }
        
        generator = load_generators.get(self.config.pattern, self._generate_constant_load)
        generator_task = asyncio.create_task(generator(request_queue))
        
        try:
            # Wait for generator to finish or duration to expire
            if self.config.duration_seconds:
                await asyncio.wait_for(generator_task, timeout=self.config.duration_seconds)
            else:
                await generator_task
            
            # Wait for all queued requests to complete with timeout
            logger.info("🔄 Waiting for remaining requests to complete...")
            try:
                await asyncio.wait_for(request_queue.join(), timeout=10.0)
            except asyncio.TimeoutError:
                logger.warning("⚠️  Timeout waiting for queue completion, proceeding with shutdown")
            
        except asyncio.TimeoutError:
            logger.info("⏰ Load test duration expired")
        except (KeyboardInterrupt, asyncio.CancelledError):
            logger.info("🛑 Load test interrupted by user")
        finally:
            self.running = False
            
            # Cancel all tasks gracefully
            generator_task.cancel()
            reporter_task.cancel()
            for worker in workers:
                worker.cancel()
            
            # Wait briefly for cancellation to complete
            try:
                await asyncio.wait_for(
                    asyncio.gather(generator_task, reporter_task, *workers, return_exceptions=True),
                    timeout=2.0
                )
            except asyncio.TimeoutError:
                logger.warning("⚠️  Task cancellation timeout")
            
            # Close session
            if self.session:
                await self.session.close()
            
            # Final stats report
            final_stats = self.stats.get_stats_summary()
            logger.info("🏁 FINAL LOAD TEST RESULTS:")
            logger.info(f"   Total Duration: {final_stats['duration']:.1f} seconds")
            logger.info(f"   Total Requests: {final_stats['total_requests']}")
            logger.info(f"   Successful: {final_stats['completed']} ({final_stats['success_rate']:.1f}%)")
            logger.info(f"   Failed: {final_stats['failed']}")
            logger.info(f"   Average RPS: {final_stats['avg_rps']:.1f}")
            logger.info(f"   Peak Concurrency: {final_stats['max_concurrent']}")
            logger.info(f"   Average Response Time: {final_stats['avg_response_time']:.2f}s")
            
            if final_stats['percentiles']:
                p = final_stats['percentiles']
                logger.info(f"   Response Time Percentiles:")
                logger.info(f"     p50: {p['p50']:.2f}s")
                logger.info(f"     p95: {p['p95']:.2f}s") 
                logger.info(f"     p99: {p['p99']:.2f}s")
                logger.info(f"     max: {p['max']:.2f}s")

async def main():
    """Main entry point"""
    parser = argparse.ArgumentParser(description='High-Performance Ollama Load Tester')
    parser.add_argument('--url', default='http://localhost:11435', 
                       help='Base URL (default: monitoring proxy)')
    parser.add_argument('--model', default='phi3:mini', help='Model name')
    parser.add_argument('--pattern', choices=['constant', 'burst', 'ramp', 'spike', 'chaos'],
                       default='constant', help='Load pattern')
    parser.add_argument('--rps', type=float, default=2.0, help='Target requests per second')
    parser.add_argument('--concurrent', type=int, default=5, help='Max concurrent requests')
    parser.add_argument('--requests', type=int, default=100, help='Total requests to send')
    parser.add_argument('--duration', type=int, help='Test duration in seconds (optional)')
    parser.add_argument('--burst-size', type=int, default=20, help='Requests per burst')
    parser.add_argument('--burst-interval', type=float, default=5.0, help='Seconds between bursts')
    parser.add_argument('--prompts', nargs='+', choices=['short', 'medium', 'long'],
                       default=['short', 'medium', 'long'], help='Prompt types to use')
    
    args = parser.parse_args()
    
    config = LoadTestConfig(
        base_url=args.url,
        model=args.model,
        pattern=args.pattern,
        requests_per_second=args.rps,
        max_concurrent_requests=args.concurrent,
        total_requests=args.requests,
        duration_seconds=args.duration,
        burst_size=args.burst_size,
        burst_interval=args.burst_interval,
        prompt_types=args.prompts
    )
    
    load_tester = HighPerformanceLoadTester(config)
    await load_tester.run_load_test()

if __name__ == "__main__":
    asyncio.run(main())