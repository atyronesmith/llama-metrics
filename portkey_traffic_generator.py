#!/usr/bin/env python3
"""
Portkey Traffic Generator for Monitoring

This program generates traffic through Portkey Gateway to Ollama for monitoring
Portkey's routing, caching, and fallback capabilities.
"""

import asyncio
import random
import time
import json
import os
import sys
from typing import List, Dict, Optional
from datetime import datetime
import aiohttp
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler('portkey_traffic.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class PortkeyTrafficGenerator:
    def __init__(self, model: str = "phi3:mini", 
                 portkey_url: str = "http://localhost:8787",
                 proxy_url: str = "http://localhost:11435"):
        self.model = model
        self.portkey_url = portkey_url
        self.proxy_url = proxy_url
        self.questions = []
        self.stats = {
            "total_requests": 0,
            "successful_requests": 0,
            "failed_requests": 0,
            "total_latency": 0.0,
            "portkey_requests": 0,
            "proxy_requests": 0,
            "start_time": None,
            "last_request_time": None
        }

    def load_questions(self) -> None:
        """Load questions from all JSON files in the questions directory"""
        self.questions = []
        questions_dir = "questions"
        
        if not os.path.exists(questions_dir):
            logger.error(f"Questions directory '{questions_dir}' not found!")
            logger.info("Creating sample questions...")
            self.create_sample_questions()
            return

        try:
            for filename in os.listdir(questions_dir):
                if filename.endswith('.json'):
                    filepath = os.path.join(questions_dir, filename)
                    with open(filepath, 'r') as f:
                        data = json.load(f)
                        if 'questions' in data:
                            self.questions.extend(data['questions'])
                        elif isinstance(data, list):
                            self.questions.extend(data)
        except Exception as e:
            logger.error(f"Error loading questions: {e}")
            self.create_sample_questions()

        if self.questions:
            logger.info(f"Loaded {len(self.questions)} questions from {questions_dir}")
        else:
            self.create_sample_questions()

    def create_sample_questions(self) -> None:
        """Create sample questions for testing"""
        self.questions = [
            "What is artificial intelligence?",
            "Explain the concept of machine learning.",
            "How does deep learning work?",
            "What are neural networks?",
            "Describe the difference between supervised and unsupervised learning.",
            "What is natural language processing?",
            "How do large language models work?",
            "What are the benefits of using AI in healthcare?",
            "Explain the concept of computer vision.",
            "What are the ethical considerations in AI development?",
            "How does reinforcement learning work?",
            "What is the difference between AI, ML, and deep learning?",
            "Describe the process of training a neural network.",
            "What are transformer models?",
            "How does attention mechanism work in transformers?",
            "What is transfer learning?",
            "Explain the concept of overfitting in machine learning.",
            "What are GANs and how do they work?",
            "Describe the applications of AI in autonomous vehicles.",
            "What is the role of data in machine learning?"
        ]
        logger.info(f"Created {len(self.questions)} sample questions")

    async def send_portkey_request(self, question: str) -> Optional[Dict]:
        """Send request directly to Portkey Gateway"""
        payload = {
            "model": self.model,
            "messages": [{"role": "user", "content": question}],
            "stream": False,
            "max_tokens": 300
        }

        headers = {
            "Content-Type": "application/json",
            # Portkey will handle routing to Ollama based on portkey-config.json
        }

        start_time = time.time()
        
        try:
            async with aiohttp.ClientSession() as session:
                async with session.post(
                    f"{self.portkey_url}/v1/chat/completions",
                    json=payload,
                    headers=headers,
                    timeout=aiohttp.ClientTimeout(total=120)
                ) as response:
                    
                    response_data = await response.json()
                    latency = time.time() - start_time
                    
                    self.stats["portkey_requests"] += 1
                    
                    if response.status == 200:
                        logger.info(f"✅ Portkey request successful (Status: {response.status}, Latency: {latency:.2f}s)")
                        return {"status": "success", "latency": latency, "response": response_data}
                    else:
                        logger.warning(f"⚠️  Portkey request failed (Status: {response.status})")
                        return {"status": "failed", "latency": latency, "error": response_data}

        except asyncio.TimeoutError:
            latency = time.time() - start_time
            logger.error("❌ Portkey request timeout")
            return {"status": "timeout", "latency": latency}
        except Exception as e:
            latency = time.time() - start_time
            logger.error(f"❌ Portkey request error: {e}")
            return {"status": "error", "latency": latency, "error": str(e)}

    async def send_proxy_request(self, question: str) -> Optional[Dict]:
        """Send request to monitoring proxy with Portkey enabled"""
        payload = {
            "model": self.model,
            "prompt": question,
            "stream": False
        }

        headers = {
            "Content-Type": "application/json"
        }

        start_time = time.time()
        
        try:
            async with aiohttp.ClientSession() as session:
                async with session.post(
                    f"{self.proxy_url}/api/generate",
                    json=payload,
                    headers=headers,
                    timeout=aiohttp.ClientTimeout(total=120)
                ) as response:
                    
                    response_data = await response.json()
                    latency = time.time() - start_time
                    
                    self.stats["proxy_requests"] += 1
                    
                    if response.status == 200:
                        logger.info(f"✅ Proxy request successful (Status: {response.status}, Latency: {latency:.2f}s)")
                        return {"status": "success", "latency": latency, "response": response_data}
                    else:
                        logger.warning(f"⚠️  Proxy request failed (Status: {response.status})")
                        return {"status": "failed", "latency": latency, "error": response_data}

        except asyncio.TimeoutError:
            latency = time.time() - start_time
            logger.error("❌ Proxy request timeout")
            return {"status": "timeout", "latency": latency}
        except Exception as e:
            latency = time.time() - start_time
            logger.error(f"❌ Proxy request error: {e}")
            return {"status": "error", "latency": latency, "error": str(e)}

    async def run_mixed_traffic(self, num_requests: int = 50, delay: float = 2.0):
        """Run mixed traffic through both Portkey direct and monitoring proxy"""
        if not self.questions:
            logger.error("No questions available!")
            return

        self.stats["start_time"] = datetime.now()
        logger.info(f"🚀 Starting mixed Portkey traffic: {num_requests} requests with {delay}s delay")
        logger.info(f"📊 Direct Portkey: {self.portkey_url}")
        logger.info(f"🔄 Monitoring Proxy: {self.proxy_url}")

        for i in range(num_requests):
            question = random.choice(self.questions)
            self.stats["total_requests"] += 1
            
            # Alternate between direct Portkey and monitoring proxy
            if i % 2 == 0:
                logger.info(f"📤 Request {i+1}/{num_requests} via Portkey: {question[:50]}...")
                result = await self.send_portkey_request(question)
            else:
                logger.info(f"📤 Request {i+1}/{num_requests} via Proxy: {question[:50]}...")
                result = await self.send_proxy_request(question)
            
            if result and result.get("status") == "success":
                self.stats["successful_requests"] += 1
                self.stats["total_latency"] += result.get("latency", 0)
            else:
                self.stats["failed_requests"] += 1
            
            self.stats["last_request_time"] = datetime.now()
            
            if i < num_requests - 1:
                await asyncio.sleep(delay)

        self.print_final_stats()

    async def run_portkey_only_traffic(self, num_requests: int = 15, delay: float = 1.5):
        """Run traffic only through Portkey Gateway directly"""
        if not self.questions:
            logger.error("No questions available!")
            return

        self.stats["start_time"] = datetime.now()
        logger.info(f"🚪 Starting Portkey-only traffic: {num_requests} requests with {delay}s delay")
        logger.info(f"📊 Direct Portkey: {self.portkey_url}")

        for i in range(num_requests):
            question = random.choice(self.questions)
            self.stats["total_requests"] += 1
            
            logger.info(f"📤 Request {i+1}/{num_requests} via Portkey: {question[:50]}...")
            result = await self.send_portkey_request(question)
            
            if result and result.get("status") == "success":
                self.stats["successful_requests"] += 1
                self.stats["total_latency"] += result.get("latency", 0)
            else:
                self.stats["failed_requests"] += 1
            
            self.stats["last_request_time"] = datetime.now()
            
            if i < num_requests - 1:
                await asyncio.sleep(delay)

        self.print_final_stats()

    async def run_proxy_only_traffic(self, num_requests: int = 15, delay: float = 1.5):
        """Run traffic only through monitoring proxy (with Portkey enabled)"""
        if not self.questions:
            logger.error("No questions available!")
            return

        self.stats["start_time"] = datetime.now()
        logger.info(f"🔄 Starting Proxy-only traffic: {num_requests} requests with {delay}s delay")
        logger.info(f"🔄 Monitoring Proxy: {self.proxy_url}")

        for i in range(num_requests):
            question = random.choice(self.questions)
            self.stats["total_requests"] += 1
            
            logger.info(f"📤 Request {i+1}/{num_requests} via Proxy: {question[:50]}...")
            result = await self.send_proxy_request(question)
            
            if result and result.get("status") == "success":
                self.stats["successful_requests"] += 1
                self.stats["total_latency"] += result.get("latency", 0)
            else:
                self.stats["failed_requests"] += 1
            
            self.stats["last_request_time"] = datetime.now()
            
            if i < num_requests - 1:
                await asyncio.sleep(delay)

        self.print_final_stats()

    def print_final_stats(self):
        """Print final statistics"""
        total_time = (self.stats["last_request_time"] - self.stats["start_time"]).total_seconds()
        avg_latency = self.stats["total_latency"] / max(self.stats["successful_requests"], 1)
        success_rate = (self.stats["successful_requests"] / self.stats["total_requests"]) * 100

        print("\n" + "="*60)
        print("🎯 PORTKEY TRAFFIC GENERATOR RESULTS")
        print("="*60)
        print(f"📊 Total Requests:     {self.stats['total_requests']}")
        print(f"✅ Successful:         {self.stats['successful_requests']}")
        print(f"❌ Failed:             {self.stats['failed_requests']}")
        print(f"🚪 Direct Portkey:     {self.stats['portkey_requests']}")
        print(f"🔄 Via Proxy:          {self.stats['proxy_requests']}")
        print(f"⏱️  Average Latency:    {avg_latency:.2f}s")
        print(f"📈 Success Rate:       {success_rate:.1f}%")
        print(f"⌚ Total Time:         {total_time:.1f}s")
        print("="*60)

async def main():
    import argparse
    
    parser = argparse.ArgumentParser(description="Portkey Traffic Generator")
    parser.add_argument("--model", default="phi3:mini", help="Model name")
    parser.add_argument("--portkey-url", default="http://localhost:8787", 
                       help="Portkey Gateway URL")
    parser.add_argument("--proxy-url", default="http://localhost:11435", 
                       help="Monitoring proxy URL") 
    parser.add_argument("--requests", type=int, default=20, 
                       help="Number of requests to send")
    parser.add_argument("--delay", type=float, default=2.0, 
                       help="Delay between requests in seconds")
    parser.add_argument("--mode", choices=["mixed", "portkey", "proxy"], default="mixed",
                       help="Traffic mode: mixed (default), portkey only, or proxy only")
    
    args = parser.parse_args()
    
    generator = PortkeyTrafficGenerator(
        model=args.model,
        portkey_url=args.portkey_url,
        proxy_url=args.proxy_url
    )
    
    generator.load_questions()
    
    if args.mode == "mixed":
        await generator.run_mixed_traffic(args.requests, args.delay)
    elif args.mode == "portkey":
        await generator.run_portkey_only_traffic(args.requests, args.delay)
    elif args.mode == "proxy":
        await generator.run_proxy_only_traffic(args.requests, args.delay)

if __name__ == "__main__":
    asyncio.run(main())