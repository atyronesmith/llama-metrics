#!/usr/bin/env python3
"""
OpenAI Compatibility Test for Ollama Monitoring Proxy

This script demonstrates using the OpenAI Python SDK with the Ollama proxy.
It tests both streaming and non-streaming chat completions.

Requirements:
    pip install openai
"""

import os
import time
import json
from openai import OpenAI

# Configure the client to use the proxy
client = OpenAI(
    api_key="not-needed",  # Ollama doesn't require API keys
    base_url="http://localhost:11434/v1"  # Proxy URL with OpenAI path
)

def test_non_streaming_chat():
    """Test non-streaming chat completion"""
    print("\n=== Testing Non-Streaming Chat Completion ===")

    start_time = time.time()

    try:
        response = client.chat.completions.create(
            model="gpt-3.5-turbo",  # Will be mapped to llama2:13b
            messages=[
                {"role": "system", "content": "You are a helpful assistant."},
                {"role": "user", "content": "What is the capital of France? Answer in one sentence."}
            ],
            max_tokens=50,
            temperature=0.7,
            user="test-user-123"  # For user tracking
        )

        end_time = time.time()

        print(f"Model: {response.model}")
        print(f"Response: {response.choices[0].message.content}")
        print(f"Finish Reason: {response.choices[0].finish_reason}")
        print(f"Usage: {response.usage}")
        print(f"Response Time: {end_time - start_time:.2f}s")
        print(f"Request ID: {response.id}")

    except Exception as e:
        print(f"Error: {e}")

def test_streaming_chat():
    """Test streaming chat completion"""
    print("\n=== Testing Streaming Chat Completion ===")

    start_time = time.time()
    first_token_time = None
    full_response = ""

    try:
        response = client.chat.completions.create(
            model="gpt-3.5-turbo",  # Will be mapped to llama2:13b
            messages=[
                {"role": "system", "content": "You are a helpful assistant."},
                {"role": "user", "content": "Write a haiku about monitoring AI systems."}
            ],
            stream=True,
            temperature=0.7,
            user="test-user-456"  # For user tracking
        )

        print("Streaming response:")
        for chunk in response:
            if chunk.choices[0].delta.content:
                if first_token_time is None:
                    first_token_time = time.time()
                    print(f"Time to first token: {first_token_time - start_time:.3f}s")

                content = chunk.choices[0].delta.content
                print(content, end="", flush=True)
                full_response += content

        end_time = time.time()

        print(f"\n\nStreaming completed")
        print(f"Total Response Time: {end_time - start_time:.2f}s")
        print(f"Response Length: {len(full_response)} characters")

    except Exception as e:
        print(f"Error: {e}")

def test_model_listing():
    """Test model listing endpoint"""
    print("\n=== Testing Model Listing ===")

    try:
        models = client.models.list()
        print("Available models:")
        for model in models:
            print(f"  - {model.id}")
    except Exception as e:
        print(f"Error listing models: {e}")

def test_legacy_completion():
    """Test legacy completion API"""
    print("\n=== Testing Legacy Completion API ===")

    try:
        response = client.completions.create(
            model="text-davinci-003",  # Will be mapped to llama2:7b
            prompt="The capital of France is",
            max_tokens=10,
            temperature=0.5
        )

        print(f"Model: {response.model}")
        print(f"Response: {response.choices[0].text}")
        print(f"Usage: {response.usage}")

    except Exception as e:
        print(f"Error: {e}")

def check_metrics():
    """Check if metrics are being collected"""
    print("\n=== Checking Metrics Endpoint ===")

    import requests

    try:
        response = requests.get("http://localhost:9090/metrics")
        if response.status_code == 200:
            print("Metrics endpoint is accessible")

            # Look for our custom metrics
            metrics_to_check = [
                "ollama_proxy_requests_total",
                "ollama_proxy_tokens_per_second",
                "ollama_proxy_time_to_first_token_seconds",
                "ollama_proxy_user_requests_total",
                "ollama_proxy_token_cost_total"
            ]

            metrics_text = response.text
            for metric in metrics_to_check:
                if metric in metrics_text:
                    print(f"✓ Found metric: {metric}")
                else:
                    print(f"✗ Missing metric: {metric}")
        else:
            print(f"Metrics endpoint returned status code: {response.status_code}")
    except Exception as e:
        print(f"Error checking metrics: {e}")

def main():
    """Run all tests"""
    print("OpenAI Compatibility Test Suite for Ollama Monitoring Proxy")
    print("=" * 60)

    # Check if proxy is running
    try:
        test_model_listing()
    except Exception as e:
        print(f"\nError: Cannot connect to proxy. Is it running on http://localhost:11434?")
        print(f"Details: {e}")
        return

    # Run tests
    test_non_streaming_chat()
    time.sleep(1)  # Brief pause between tests

    test_streaming_chat()
    time.sleep(1)

    test_legacy_completion()
    time.sleep(1)

    check_metrics()

    print("\n" + "=" * 60)
    print("All tests completed!")
    print("\nYou can view detailed metrics at http://localhost:9090/metrics")
    print("Look for metrics with the 'ollama_proxy_' prefix")

if __name__ == "__main__":
    main()