#!/usr/bin/env python3
"""Test script for priority queue functionality in Ollama proxy."""

import requests
import json
import time
import threading
from datetime import datetime

PROXY_URL = "http://localhost:11435"

def send_request(priority, request_id, prompt):
    """Send a request with specified priority."""
    headers = {
        "Content-Type": "application/json"
    }

    if priority == "high":
        headers["X-Priority"] = "high"

    data = {
        "model": "phi3:mini",
        "prompt": f"[{priority.upper()} PRIORITY {request_id}] {prompt}",
        "stream": False,
        "options": {
            "num_predict": 10  # Short response for testing
        }
    }

    start_time = time.time()
    print(f"[{datetime.now().strftime('%H:%M:%S.%f')[:-3]}] Sending {priority} priority request {request_id}")

    try:
        response = requests.post(
            f"{PROXY_URL}/api/generate",
            headers=headers,
            json=data,
            timeout=60
        )

        elapsed = time.time() - start_time

        if response.status_code == 200:
            result = response.json()
            print(f"[{datetime.now().strftime('%H:%M:%S.%f')[:-3]}] Completed {priority} priority request {request_id} in {elapsed:.2f}s")
            return True
        else:
            print(f"[{datetime.now().strftime('%H:%M:%S.%f')[:-3]}] Failed {priority} priority request {request_id}: {response.status_code}")
            return False

    except Exception as e:
        elapsed = time.time() - start_time
        print(f"[{datetime.now().strftime('%H:%M:%S.%f')[:-3]}] Error {priority} priority request {request_id} after {elapsed:.2f}s: {e}")
        return False

def test_priority_queue():
    """Test the priority queue by sending mixed priority requests."""
    print("\n=== Testing Priority Queue ===")
    print("Sending 10 normal priority requests followed by 5 high priority requests...")
    print("High priority requests should be processed before normal ones.\n")

    threads = []

    # Send normal priority requests first
    for i in range(1, 11):
        t = threading.Thread(target=send_request, args=("normal", i, "What is 2+2?"))
        threads.append(t)
        t.start()
        time.sleep(0.1)  # Small delay between requests

    # Wait a bit then send high priority requests
    time.sleep(1)
    print("\n--- Now sending HIGH PRIORITY requests ---\n")

    for i in range(1, 6):
        t = threading.Thread(target=send_request, args=("high", i, "What is 1+1?"))
        threads.append(t)
        t.start()
        time.sleep(0.1)

    # Wait for all threads to complete
    for t in threads:
        t.join()

    print("\n=== Test Complete ===")
    print("\nNote: Check the proxy logs to see the queue metrics and processing order.")
    print("High priority requests should have been processed before normal priority ones.")

def check_metrics():
    """Check queue metrics from Prometheus endpoint."""
    try:
        response = requests.get("http://localhost:8001/metrics")
        if response.status_code == 200:
            lines = response.text.split('\n')
            print("\n=== Queue Metrics ===")
            for line in lines:
                if 'ollama_proxy_queue' in line and not line.startswith('#'):
                    print(line)
    except Exception as e:
        print(f"Failed to get metrics: {e}")

if __name__ == "__main__":
    # Check if proxy is running
    try:
        response = requests.get(f"{PROXY_URL}/api/tags", timeout=5)
        if response.status_code != 200:
            print("Error: Proxy is not responding correctly")
            exit(1)
    except:
        print("Error: Cannot connect to proxy at", PROXY_URL)
        print("Make sure the proxy is running: make start-proxy")
        exit(1)

    test_priority_queue()
    time.sleep(2)
    check_metrics()