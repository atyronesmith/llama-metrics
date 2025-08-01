#!/usr/bin/env python3
"""
Mac System Metrics Helper
Collects system metrics that require elevated permissions and exposes them via HTTP
"""

import subprocess
import json
import time
import threading
from flask import Flask, jsonify
import logging
import re

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = Flask(__name__)

# Global metrics storage
metrics = {
    'gpu_utilization': 0.0,
    'gpu_power': 0.0,
    'cpu_power': 0.0,
    'cpu_temperature': 0.0,
    'memory_pressure': 0.0,
    'thermal_pressure': 'nominal',
    'timestamp': time.time()
}

def run_command(cmd):
    """Run a shell command and return output"""
    try:
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True, timeout=5)
        if result.returncode == 0:
            return result.stdout
        else:
            logger.error(f"Command failed: {cmd}, Error: {result.stderr}")
            return None
    except Exception as e:
        logger.error(f"Error running command {cmd}: {e}")
        return None

def collect_powermetrics():
    """Collect metrics using powermetrics (requires sudo)"""
    global metrics

    # Run powermetrics for 1 second sample
    cmd = "sudo powermetrics --samplers gpu_power,cpu_power,thermal --sample-count 1 --sample-rate 1000 -f json"
    output = run_command(cmd)

    if output:
        try:
            data = json.loads(output)

            # GPU metrics
            if 'gpu' in data:
                gpu_data = data['gpu']
                if 'idle_ratio' in gpu_data:
                    metrics['gpu_utilization'] = (1.0 - gpu_data['idle_ratio']) * 100
                if 'power' in gpu_data:
                    metrics['gpu_power'] = gpu_data['power']

            # CPU power
            if 'processor' in data:
                proc_data = data['processor']
                if 'package_power' in proc_data:
                    metrics['cpu_power'] = proc_data['package_power']

            # Thermal pressure
            if 'thermal_pressure' in data:
                metrics['thermal_pressure'] = data['thermal_pressure']

        except json.JSONDecodeError as e:
            logger.error(f"Error parsing powermetrics JSON: {e}")

def collect_temperature():
    """Collect CPU temperature"""
    global metrics

    # Try osx-cpu-temp first
    output = run_command("osx-cpu-temp")
    if output:
        # Parse output like "45.5°C"
        match = re.search(r'([\d.]+)°C', output)
        if match:
            metrics['cpu_temperature'] = float(match.group(1))
            return

    # Fallback to powermetrics SMC data
    cmd = "sudo powermetrics --samplers smc -n 1 -i 1000"
    output = run_command(cmd)

    if output:
        # Look for CPU die temperature
        match = re.search(r'CPU die temperature:\s+([\d.]+)\s+C', output)
        if match:
            metrics['cpu_temperature'] = float(match.group(1))

def collect_memory_pressure():
    """Collect memory pressure statistics"""
    global metrics

    output = run_command("memory_pressure")
    if output:
        # Parse memory pressure output
        match = re.search(r'System-wide memory free percentage:\s+([\d.]+)%', output)
        if match:
            free_percent = float(match.group(1))
            metrics['memory_pressure'] = 100 - free_percent

def collect_metrics_loop():
    """Background thread to collect metrics"""
    while True:
        try:
            collect_powermetrics()
            collect_temperature()
            collect_memory_pressure()
            metrics['timestamp'] = time.time()
        except Exception as e:
            logger.error(f"Error in metrics collection: {e}")

        # Collect every 5 seconds
        time.sleep(5)

@app.route('/metrics')
def get_metrics():
    """Return current metrics in Prometheus format"""
    prometheus_output = []
    
    # Add help and type information for each metric
    prometheus_output.extend([
        "# HELP mac_gpu_utilization_percent GPU utilization percentage",
        "# TYPE mac_gpu_utilization_percent gauge",
        f"mac_gpu_utilization_percent {metrics['gpu_utilization']}",
        "",
        "# HELP mac_gpu_power_watts GPU power consumption in watts", 
        "# TYPE mac_gpu_power_watts gauge",
        f"mac_gpu_power_watts {metrics['gpu_power']}",
        "",
        "# HELP mac_cpu_power_watts CPU power consumption in watts",
        "# TYPE mac_cpu_power_watts gauge", 
        f"mac_cpu_power_watts {metrics['cpu_power']}",
        "",
        "# HELP mac_cpu_temperature_celsius CPU temperature in Celsius",
        "# TYPE mac_cpu_temperature_celsius gauge",
        f"mac_cpu_temperature_celsius {metrics['cpu_temperature']}",
        "",
        "# HELP mac_memory_pressure_percent Memory pressure percentage",
        "# TYPE mac_memory_pressure_percent gauge",
        f"mac_memory_pressure_percent {metrics['memory_pressure']}",
        "",
        "# HELP mac_thermal_pressure_info Thermal pressure state (0=nominal, 1=fair, 2=serious, 3=critical)",
        "# TYPE mac_thermal_pressure_info gauge",
    ])
    
    # Convert thermal pressure to numeric value
    thermal_values = {"nominal": 0, "fair": 1, "serious": 2, "critical": 3}
    thermal_numeric = thermal_values.get(metrics['thermal_pressure'], 0)
    prometheus_output.append(f"mac_thermal_pressure_info {thermal_numeric}")
    prometheus_output.append("")
    
    # Add timestamp as a metric
    prometheus_output.extend([
        "# HELP mac_metrics_last_updated_timestamp Unix timestamp of last metrics update", 
        "# TYPE mac_metrics_last_updated_timestamp gauge",
        f"mac_metrics_last_updated_timestamp {metrics['timestamp']}",
        ""
    ])
    
    response = "\n".join(prometheus_output)
    return response, 200, {'Content-Type': 'text/plain; version=0.0.4; charset=utf-8'}

@app.route('/metrics/json')  
def get_metrics_json():
    """Return current metrics as JSON (legacy endpoint)"""
    return jsonify(metrics)

@app.route('/health')
def health():
    """Health check endpoint"""
    return jsonify({
        'status': 'healthy',
        'uptime': time.time() - start_time
    })

if __name__ == '__main__':
    start_time = time.time()

    # Check if running with sudo
    output = run_command("whoami")
    if output and output.strip() != 'root':
        logger.warning("Not running as root. Some metrics may not be available.")
        logger.warning("Run with: sudo python3 mac_metrics_helper.py")

    # Start metrics collection thread
    collector_thread = threading.Thread(target=collect_metrics_loop, daemon=True)
    collector_thread.start()

    # Start Flask server
    logger.info("Starting Mac metrics helper on http://localhost:8002")
    app.run(host='0.0.0.0', port=8002, debug=False)