#!/usr/bin/env python3
"""
Ollama Traffic Generator for Monitoring

This program generates continuous prompt traffic to Ollama to help monitor
model performance and collect metrics. It asks questions from a curated
list of 1000 diverse questions, processing them serially with random selection.
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
        logging.FileHandler('traffic_generator.log'),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger(__name__)

class OllamaTrafficGenerator:
    def __init__(self, model: str = "phi3:mini", base_url: str = "http://localhost:8000"):
        self.model = model
        self.base_url = base_url
        self.api_key = "sk-1234567890abcdef"  # LiteLLM API key
        self.questions = []
        self.stats = {
            "total_requests": 0,
            "successful_requests": 0,
            "failed_requests": 0,
            "total_latency": 0.0,
            "start_time": None,
            "last_request_time": None
        }

    def load_questions(self) -> None:
        """Load questions from all JSON files in the questions directory"""
        self.questions = []
        questions_dir = "questions"

        if not os.path.exists(questions_dir):
            logger.error(f"Questions directory '{questions_dir}' not found.")
            sys.exit(1)

        try:
            # Get all JSON files in the questions directory
            json_files = [f for f in os.listdir(questions_dir) if f.endswith('.json')]

            if not json_files:
                logger.error(f"No JSON files found in '{questions_dir}' directory.")
                sys.exit(1)

            categories_loaded = []

            for json_file in json_files:
                file_path = os.path.join(questions_dir, json_file)
                try:
                    with open(file_path, 'r', encoding='utf-8') as f:
                        data = json.load(f)
                        if 'questions' in data and 'category' in data:
                            self.questions.extend(data['questions'])
                            categories_loaded.append(data['category'])
                            logger.info(f"Loaded {len(data['questions'])} questions from {data['category']}")
                        else:
                            logger.warning(f"Invalid format in {json_file} - missing 'questions' or 'category' field")
                except json.JSONDecodeError as e:
                    logger.error(f"Invalid JSON in {json_file}: {e}")
                except Exception as e:
                    logger.error(f"Error loading {json_file}: {e}")

            logger.info(f"📚 Total questions loaded: {len(self.questions)}")
            logger.info(f"📁 Categories: {', '.join(categories_loaded)}")

            if not self.questions:
                logger.error("No questions were loaded successfully.")
                sys.exit(1)

        except Exception as e:
            logger.error(f"Error accessing questions directory: {e}")
            sys.exit(1)

    async def check_ollama_health(self) -> bool:
        """Check if LiteLLM proxy is running and accessible"""
        try:
            headers = {"Authorization": f"Bearer {self.api_key}"}
            async with aiohttp.ClientSession() as session:
                async with session.get(f"{self.base_url}/v1/models", headers=headers) as response:
                    if response.status == 200:
                        logger.info("✅ LiteLLM proxy is running and accessible")
                        return True
                    else:
                        logger.error(f"❌ LiteLLM proxy returned status {response.status}")
                        return False
        except Exception as e:
            logger.error(f"❌ Cannot connect to LiteLLM proxy: {e}")
            return False

    async def send_prompt(self, question: str) -> Optional[Dict]:
        """Send a single prompt to LiteLLM (OpenAI-compatible API)"""
        payload = {
            "model": self.model,
            "messages": [
                {
                    "role": "user", 
                    "content": question
                }
            ],
            "stream": False,
            "max_tokens": 500
        }

        headers = {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json"
        }

        start_time = time.time()

        try:
            async with aiohttp.ClientSession() as session:
                async with session.post(
                    f"{self.base_url}/v1/chat/completions",
                    json=payload,
                    headers=headers,
                    timeout=aiohttp.ClientTimeout(total=120)
                ) as response:

                    if response.status == 200:
                        result = await response.json()
                        latency = time.time() - start_time

                        self.stats["successful_requests"] += 1
                        self.stats["total_latency"] += latency
                        self.stats["last_request_time"] = datetime.now()

                        logger.info(f"✅ Q{self.stats['total_requests']}: {question[:50]}... (Latency: {latency:.2f}s)")

                        # Extract response from OpenAI format
                        response_text = ""
                        if 'choices' in result and result['choices']:
                            response_text = result['choices'][0].get('message', {}).get('content', '')
                        
                        if response_text:
                            logger.info(f"🤖 Response: {response_text[:100]}{'...' if len(response_text) > 100 else ''}")

                        return {
                            "question": question,
                            "response": response_text,
                            "latency": latency,
                            "timestamp": datetime.now().isoformat()
                        }
                    else:
                        error_text = await response.text()
                        logger.error(f"❌ HTTP {response.status}: {error_text}")
                        self.stats["failed_requests"] += 1
                        return None

        except asyncio.TimeoutError:
            logger.error(f"❌ Timeout for question: {question[:50]}...")
            self.stats["failed_requests"] += 1
            return None
        except Exception as e:
            logger.error(f"❌ Error sending prompt: {e}")
            self.stats["failed_requests"] += 1
            return None

    def print_stats(self) -> None:
        """Print current statistics"""
        if self.stats["successful_requests"] > 0:
            avg_latency = self.stats["total_latency"] / self.stats["successful_requests"]
            success_rate = (self.stats["successful_requests"] / self.stats["total_requests"]) * 100

            logger.info(f"""
📊 Current Statistics:
   Total Requests: {self.stats['total_requests']}
   Successful: {self.stats['successful_requests']}
   Failed: {self.stats['failed_requests']}
   Success Rate: {success_rate:.1f}%
   Average Latency: {avg_latency:.2f}s
            """)

    async def run(self, max_questions: Optional[int] = None, delay: float = 1.0):
        """Main loop to generate traffic"""
        logger.info(f"🚀 Starting traffic generator for model: {self.model}")
        logger.info(f"📝 Will process {'all' if max_questions is None else max_questions} questions")
        logger.info(f"⏱️  Delay between requests: {delay}s")

        # Check Ollama health
        if not await self.check_ollama_health():
            logger.error("❌ Ollama is not available. Exiting.")
            return

        self.stats["start_time"] = datetime.now()

        try:
            while True:
                # Randomly select a question
                question = random.choice(self.questions)
                self.stats["total_requests"] += 1

                # Send the prompt
                result = await self.send_prompt(question)

                # Print stats every 10 requests
                if self.stats["total_requests"] % 10 == 0:
                    self.print_stats()

                # Check if we've reached the limit
                if max_questions and self.stats["total_requests"] >= max_questions:
                    logger.info(f"✅ Reached limit of {max_questions} questions. Stopping.")
                    break

                # Wait before next request
                if delay > 0:
                    await asyncio.sleep(delay)

        except KeyboardInterrupt:
            logger.info("🛑 Received interrupt signal. Stopping traffic generator.")
        finally:
            self.print_stats()
            logger.info("✅ Traffic generator stopped.")

async def main():
    """Main entry point"""
    import argparse

    parser = argparse.ArgumentParser(description="Generate traffic to Ollama for monitoring")
    parser.add_argument("--model", default="phi3:mini", help="Ollama model to use")
    parser.add_argument("--url", default="http://localhost:11434", help="Ollama API URL")
    parser.add_argument("--max", type=int, help="Maximum number of questions to ask")
    parser.add_argument("--delay", type=float, default=1.0, help="Delay between requests in seconds")

    args = parser.parse_args()

    generator = OllamaTrafficGenerator(model=args.model, base_url=args.url)
    generator.load_questions()

    await generator.run(max_questions=args.max, delay=args.delay)

if __name__ == "__main__":
    asyncio.run(main())