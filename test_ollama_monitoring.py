#!/usr/bin/env python3
"""
Test script to demonstrate Ollama monitoring
"""

import asyncio
import aiohttp
import json
import time
import random

async def test_ollama_request(session, proxy_url, prompt, model="phi3:mini"):
    """Send a test request through the monitoring proxy"""
    print(f"\n📝 Testing with prompt: '{prompt[:50]}...'")

    url = f"{proxy_url}/api/generate"
    payload = {
        "model": model,
        "prompt": prompt,
        "stream": False
    }

    start_time = time.time()

    try:
        async with session.post(url, json=payload) as response:
            if response.status == 200:
                data = await response.json()
                duration = time.time() - start_time

                print(f"✅ Success! Response time: {duration:.2f}s")
                print(f"📊 Response preview: {data.get('response', '')[:100]}...")

                # Extract performance metrics if available
                if 'eval_count' in data:
                    eval_duration = data.get('eval_duration', 0) / 1e9
                    tokens_per_sec = data['eval_count'] / eval_duration if eval_duration > 0 else 0
                    print(f"⚡ Token generation rate: {tokens_per_sec:.2f} tokens/s")
                    print(f"📏 Total tokens: {data['eval_count']}")

                return True
            else:
                print(f"❌ Error: HTTP {response.status}")
                return False

    except Exception as e:
        print(f"❌ Request failed: {e}")
        return False

async def test_streaming_request(session, proxy_url, prompt, model="phi3:mini"):
    """Test streaming request through the proxy"""
    print(f"\n🌊 Testing streaming with prompt: '{prompt[:50]}...'")

    url = f"{proxy_url}/api/generate"
    payload = {
        "model": model,
        "prompt": prompt,
        "stream": True
    }

    start_time = time.time()
    first_token_time = None
    tokens = []

    try:
        async with session.post(url, json=payload) as response:
            if response.status == 200:
                async for line in response.content:
                    if line:
                        try:
                            data = json.loads(line)
                            if 'response' in data:
                                if first_token_time is None:
                                    first_token_time = time.time()
                                    ttft = first_token_time - start_time
                                    print(f"⏱️  Time to first token: {ttft:.3f}s")

                                tokens.append(data['response'])

                            if data.get('done'):
                                duration = time.time() - start_time
                                print(f"✅ Streaming complete! Total time: {duration:.2f}s")
                                print(f"📝 Generated text: {''.join(tokens)[:100]}...")
                                return True

                        except json.JSONDecodeError:
                            pass
            else:
                print(f"❌ Error: HTTP {response.status}")
                return False

    except Exception as e:
        print(f"❌ Streaming failed: {e}")
        return False

async def check_metrics(metrics_url):
    """Check if metrics are being collected"""
    print("\n📊 Checking metrics endpoint...")

    async with aiohttp.ClientSession() as session:
        try:
            async with session.get(metrics_url) as response:
                if response.status == 200:
                    text = await response.text()

                    # Check for key metrics
                    metrics_found = []
                    key_metrics = [
                        'ollama_proxy_requests_total',
                        'ollama_proxy_request_duration_seconds',
                        'ollama_proxy_tokens_per_second',
                        'ollama_proxy_active_requests',
                        'ollama_proxy_cpu_usage_percent'
                    ]

                    for metric in key_metrics:
                        if metric in text:
                            metrics_found.append(metric)

                    print(f"✅ Metrics endpoint is working!")
                    print(f"📈 Found {len(metrics_found)} key metrics:")
                    for metric in metrics_found:
                        print(f"   - {metric}")

                    return True
                else:
                    print(f"❌ Metrics endpoint returned: {response.status}")
                    return False

        except Exception as e:
            print(f"❌ Failed to check metrics: {e}")
            return False

async def run_load_test(proxy_url, num_requests=5):
    """Run a simple load test"""
    print(f"\n🔄 Running load test with {num_requests} requests...")

    prompts = [
        "What is the capital of France?",
        "Explain quantum computing in simple terms.",
        "Write a haiku about technology.",
        "What are the benefits of exercise?",
        "How do neural networks work?",
        "What is the meaning of life?",
        "Describe the water cycle.",
        "What is machine learning?"
    ]

    async with aiohttp.ClientSession() as session:
        tasks = []
        for i in range(num_requests):
            prompt = random.choice(prompts)
            task = test_ollama_request(session, proxy_url, prompt)
            tasks.append(task)
            await asyncio.sleep(0.5)  # Stagger requests

        results = await asyncio.gather(*tasks)
        success_count = sum(1 for r in results if r)

        print(f"\n📊 Load test results:")
        print(f"   ✅ Successful: {success_count}/{num_requests}")
        print(f"   ❌ Failed: {num_requests - success_count}/{num_requests}")

        return success_count == num_requests

async def main():
    """Main test function"""
    proxy_url = "http://localhost:11435"
    metrics_url = "http://localhost:8001/metrics"

    print("🚀 Ollama Monitoring Test Suite")
    print("================================")
    print(f"🔗 Proxy URL: {proxy_url}")
    print(f"📊 Metrics URL: {metrics_url}")

    # Check if services are running
    print("\n🔍 Checking service availability...")

    async with aiohttp.ClientSession() as session:
        # Test 1: Basic request
        await test_ollama_request(
            session,
            proxy_url,
            "Hello, how are you today?"
        )

        # Test 2: Streaming request
        await test_streaming_request(
            session,
            proxy_url,
            "Tell me a short story about a robot."
        )

        # Test 3: Check metrics
        await check_metrics(metrics_url)

        # Test 4: Load test
        await run_load_test(proxy_url, num_requests=3)

        # Test 5: Check metrics again
        print("\n📊 Final metrics check...")
        await check_metrics(metrics_url)

    print("\n✨ Test suite complete!")
    print("Visit the following URLs to see the results:")
    print(f"   - Metrics: {metrics_url}")
    print(f"   - Prometheus: http://localhost:9090")
    print(f"   - Grafana: http://localhost:3000")

if __name__ == "__main__":
    asyncio.run(main())