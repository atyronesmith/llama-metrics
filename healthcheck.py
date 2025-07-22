"""Health check endpoints and monitoring for Ollama Monitoring Stack."""

import asyncio
import time
import json
import logging
import psutil
import requests
from typing import Dict, Any, Optional, List
from dataclasses import dataclass, asdict
from datetime import datetime, timezone
from config_manager import get_config
from version import __version__, BUILD_INFO

# Setup logging
logger = logging.getLogger(__name__)

@dataclass
class HealthStatus:
    """Health status data structure."""
    status: str  # healthy, degraded, unhealthy
    timestamp: str
    response_time_ms: Optional[float] = None
    error: Optional[str] = None
    details: Optional[Dict[str, Any]] = None

@dataclass
class ServiceHealth:
    """Individual service health status."""
    name: str
    url: str
    status: HealthStatus
    critical: bool = True

@dataclass
class SystemHealth:
    """Overall system health status."""
    status: str
    timestamp: str
    version: str
    uptime_seconds: float
    services: List[ServiceHealth]
    system_metrics: Dict[str, Any]
    summary: Dict[str, Any]

class HealthChecker:
    """Comprehensive health checking system."""
    
    def __init__(self):
        """Initialize health checker."""
        self.config = get_config()
        self.start_time = time.time()
        
        # Service endpoints to check
        self.service_endpoints = [
            {
                "name": "ollama",
                "url": f"{self.config.server.ollama_url}/api/tags",
                "critical": True,
                "timeout": 5
            },
            {
                "name": "proxy",
                "url": f"http://{self.config.server.proxy_host}:{self.config.server.proxy_port}/health",
                "critical": True,
                "timeout": 3
            },
            {
                "name": "metrics",
                "url": f"http://{self.config.server.metrics_host}:{self.config.server.metrics_port}/metrics",
                "critical": False,
                "timeout": 3
            },
            {
                "name": "dashboard",
                "url": f"http://{self.config.server.dashboard_host}:{self.config.server.dashboard_port}/api/status",
                "critical": False,
                "timeout": 3
            }
        ]
    
    async def _check_ollama_generation_health(self) -> ServiceHealth:
        """Comprehensive Ollama health check including generation capability."""
        start_time = time.time()
        
        try:
            # First, check if Ollama is listening
            response = requests.get(
                f"{self.config.server.ollama_url}/api/tags",
                timeout=3,
                headers={"User-Agent": "HealthChecker/1.0"}
            )
            
            if response.status_code != 200:
                return ServiceHealth(
                    name="ollama",
                    url=self.config.server.ollama_url,
                    status=HealthStatus(
                        status="unhealthy",
                        timestamp=datetime.now(timezone.utc).isoformat(),
                        error=f"API endpoint not responding: HTTP {response.status_code}"
                    ),
                    critical=True
                )
            
            # Test actual generation capability with a minimal request
            gen_start = time.time()
            gen_response = requests.post(
                f"{self.config.server.ollama_url}/api/generate",
                json={
                    "model": self.config.models.default_model,
                    "prompt": "Hi",
                    "stream": False,
                    "options": {"num_predict": 1}  # Minimal generation
                },
                timeout=10,  # Reasonable timeout for generation
                headers={"Content-Type": "application/json", "User-Agent": "HealthChecker/1.0"}
            )
            
            generation_time = (time.time() - gen_start) * 1000
            total_time = (time.time() - start_time) * 1000
            
            if gen_response.status_code != 200:
                return ServiceHealth(
                    name="ollama",
                    url=self.config.server.ollama_url,
                    status=HealthStatus(
                        status="unhealthy",
                        timestamp=datetime.now(timezone.utc).isoformat(),
                        response_time_ms=round(total_time, 2),
                        error=f"Generation failed: HTTP {gen_response.status_code}",
                        details={"generation_time_ms": round(generation_time, 2)}
                    ),
                    critical=True
                )
            
            # Check if we got a valid response
            try:
                gen_data = gen_response.json()
                if not gen_data.get('response') and not gen_data.get('done'):
                    return ServiceHealth(
                        name="ollama",
                        url=self.config.server.ollama_url,
                        status=HealthStatus(
                            status="degraded",
                            timestamp=datetime.now(timezone.utc).isoformat(),
                            response_time_ms=round(total_time, 2),
                            error="Generation returned empty response",
                            details={"generation_time_ms": round(generation_time, 2)}
                        ),
                        critical=True
                    )
            except json.JSONDecodeError:
                return ServiceHealth(
                    name="ollama",
                    url=self.config.server.ollama_url,
                    status=HealthStatus(
                        status="unhealthy",
                        timestamp=datetime.now(timezone.utc).isoformat(),
                        response_time_ms=round(total_time, 2),
                        error="Generation returned invalid JSON",
                        details={"generation_time_ms": round(generation_time, 2)}
                    ),
                    critical=True
                )
            
            # All checks passed
            return ServiceHealth(
                name="ollama",
                url=self.config.server.ollama_url,
                status=HealthStatus(
                    status="healthy",
                    timestamp=datetime.now(timezone.utc).isoformat(),
                    response_time_ms=round(total_time, 2),
                    details={
                        "generation_time_ms": round(generation_time, 2),
                        "model": self.config.models.default_model,
                        "generation_working": True
                    }
                ),
                critical=True
            )
                
        except requests.exceptions.Timeout:
            return ServiceHealth(
                name="ollama",
                url=self.config.server.ollama_url,
                status=HealthStatus(
                    status="unhealthy",
                    timestamp=datetime.now(timezone.utc).isoformat(),
                    error="Request timeout - generation may be stuck"
                ),
                critical=True
            )
        except requests.exceptions.ConnectionError:
            return ServiceHealth(
                name="ollama",
                url=self.config.server.ollama_url,
                status=HealthStatus(
                    status="unhealthy",
                    timestamp=datetime.now(timezone.utc).isoformat(),
                    error="Connection refused"
                ),
                critical=True
            )
        except Exception as e:
            return ServiceHealth(
                name="ollama",
                url=self.config.server.ollama_url,
                status=HealthStatus(
                    status="unhealthy",
                    timestamp=datetime.now(timezone.utc).isoformat(),
                    error=str(e)
                ),
                critical=True
            )

    async def check_service_health(self, service: Dict[str, Any]) -> ServiceHealth:
        """Check health of a single service."""
        # Special handling for Ollama to test generation capability
        if service["name"] == "ollama":
            return await self._check_ollama_generation_health()
            
        start_time = time.time()
        
        try:
            # Make health check request
            response = requests.get(
                service["url"],
                timeout=service.get("timeout", 5),
                headers={"User-Agent": "HealthChecker/1.0"}
            )
            
            response_time = (time.time() - start_time) * 1000  # Convert to milliseconds
            
            if response.status_code == 200:
                status = HealthStatus(
                    status="healthy",
                    timestamp=datetime.now(timezone.utc).isoformat(),
                    response_time_ms=round(response_time, 2)
                )
            else:
                status = HealthStatus(
                    status="unhealthy",
                    timestamp=datetime.now(timezone.utc).isoformat(),
                    response_time_ms=round(response_time, 2),
                    error=f"HTTP {response.status_code}"
                )
                
        except requests.exceptions.ConnectTimeout:
            status = HealthStatus(
                status="unhealthy",
                timestamp=datetime.now(timezone.utc).isoformat(),
                error="Connection timeout"
            )
        except requests.exceptions.ConnectionError:
            status = HealthStatus(
                status="unhealthy",
                timestamp=datetime.now(timezone.utc).isoformat(),
                error="Connection refused"
            )
        except Exception as e:
            status = HealthStatus(
                status="unhealthy",
                timestamp=datetime.now(timezone.utc).isoformat(),
                error=str(e)
            )
        
        return ServiceHealth(
            name=service["name"],
            url=service["url"],
            status=status,
            critical=service.get("critical", True)
        )
    
    def get_system_metrics(self) -> Dict[str, Any]:
        """Get current system metrics."""
        try:
            # CPU and Memory
            cpu_percent = psutil.cpu_percent(interval=0.1)
            memory = psutil.virtual_memory()
            disk = psutil.disk_usage('/')
            
            # Network (basic)
            network = psutil.net_io_counters()
            
            metrics = {
                "cpu": {
                    "percent": round(cpu_percent, 2),
                    "count": psutil.cpu_count(),
                    "load_avg": list(psutil.getloadavg()) if hasattr(psutil, 'getloadavg') else None
                },
                "memory": {
                    "percent": round(memory.percent, 2),
                    "total_gb": round(memory.total / (1024**3), 2),
                    "available_gb": round(memory.available / (1024**3), 2),
                    "used_gb": round(memory.used / (1024**3), 2)
                },
                "disk": {
                    "percent": round(disk.percent, 2),
                    "total_gb": round(disk.total / (1024**3), 2),
                    "free_gb": round(disk.free / (1024**3), 2),
                    "used_gb": round(disk.used / (1024**3), 2)
                },
                "network": {
                    "bytes_sent": network.bytes_sent,
                    "bytes_recv": network.bytes_recv,
                    "packets_sent": network.packets_sent,
                    "packets_recv": network.packets_recv
                }
            }
            
            # macOS specific metrics
            try:
                import subprocess
                
                # GPU utilization (Metal)
                try:
                    result = subprocess.run(
                        ["system_profiler", "SPDisplaysDataType", "-json"],
                        capture_output=True, text=True, timeout=5
                    )
                    if result.returncode == 0:
                        gpu_data = json.loads(result.stdout)
                        metrics["gpu"] = {
                            "available": True,
                            "data": gpu_data.get("SPDisplaysDataType", [])
                        }
                except Exception:
                    metrics["gpu"] = {"available": False}
                
                # Power metrics (macOS)
                try:
                    result = subprocess.run(
                        ["pmset", "-g", "ps"], 
                        capture_output=True, text=True, timeout=3
                    )
                    if result.returncode == 0:
                        power_info = result.stdout
                        metrics["power"] = {
                            "available": True,
                            "battery_info": power_info
                        }
                except Exception:
                    metrics["power"] = {"available": False}
                    
            except ImportError:
                pass
            
            return metrics
            
        except Exception as e:
            logger.error(f"Failed to get system metrics: {e}")
            return {"error": str(e)}
    
    async def get_comprehensive_health(self) -> SystemHealth:
        """Get comprehensive system health status."""
        timestamp = datetime.now(timezone.utc).isoformat()
        uptime = time.time() - self.start_time
        
        # Check all services concurrently
        service_tasks = [
            self.check_service_health(service) 
            for service in self.service_endpoints
        ]
        service_results = await asyncio.gather(*service_tasks, return_exceptions=True)
        
        # Process service results
        services = []
        critical_failures = 0
        total_failures = 0
        
        for i, result in enumerate(service_results):
            if isinstance(result, Exception):
                # Create failed service health for exceptions
                service = ServiceHealth(
                    name=self.service_endpoints[i]["name"],
                    url=self.service_endpoints[i]["url"],
                    status=HealthStatus(
                        status="unhealthy",
                        timestamp=timestamp,
                        error=str(result)
                    ),
                    critical=self.service_endpoints[i].get("critical", True)
                )
            else:
                service = result
            
            services.append(service)
            
            if service.status.status != "healthy":
                total_failures += 1
                if service.critical:
                    critical_failures += 1
        
        # Determine overall health status
        if critical_failures > 0:
            overall_status = "unhealthy"
        elif total_failures > 0:
            overall_status = "degraded"
        else:
            overall_status = "healthy"
        
        # Get system metrics
        system_metrics = self.get_system_metrics()
        
        # Create summary
        healthy_services = len([s for s in services if s.status.status == "healthy"])
        summary = {
            "overall_status": overall_status,
            "services_healthy": healthy_services,
            "services_total": len(services),
            "critical_failures": critical_failures,
            "uptime_seconds": round(uptime, 2),
            "version": __version__,
            "build_info": BUILD_INFO
        }
        
        return SystemHealth(
            status=overall_status,
            timestamp=timestamp,
            version=__version__,
            uptime_seconds=round(uptime, 2),
            services=services,
            system_metrics=system_metrics,
            summary=summary
        )
    
    async def get_simple_health(self) -> Dict[str, Any]:
        """Get simple health status (fast response)."""
        timestamp = datetime.now(timezone.utc).isoformat()
        uptime = time.time() - self.start_time
        
        # Quick system check
        try:
            cpu_percent = psutil.cpu_percent(interval=0)
            memory_percent = psutil.virtual_memory().percent
            
            return {
                "status": "healthy",
                "timestamp": timestamp,
                "version": __version__,
                "uptime_seconds": round(uptime, 2),
                "system": {
                    "cpu_percent": round(cpu_percent, 2),
                    "memory_percent": round(memory_percent, 2)
                }
            }
        except Exception as e:
            return {
                "status": "degraded",
                "timestamp": timestamp,
                "version": __version__,
                "uptime_seconds": round(uptime, 2),
                "error": str(e)
            }
    
    def get_readiness_status(self) -> Dict[str, Any]:
        """Get readiness status (for container orchestration)."""
        timestamp = datetime.now(timezone.utc).isoformat()
        
        # Check critical dependencies
        ready = True
        components = {}
        
        # Check if we can access configuration
        try:
            config = get_config()
            components["config"] = "ready"
        except Exception as e:
            components["config"] = f"failed: {e}"
            ready = False
        
        # Check if we can collect system metrics
        try:
            psutil.cpu_percent()
            components["metrics_collection"] = "ready"
        except Exception as e:
            components["metrics_collection"] = f"failed: {e}"
            ready = False
        
        return {
            "ready": ready,
            "timestamp": timestamp,
            "components": components
        }
    
    def get_liveness_status(self) -> Dict[str, Any]:
        """Get liveness status (for container orchestration)."""
        timestamp = datetime.now(timezone.utc).isoformat()
        uptime = time.time() - self.start_time
        
        # Simple liveness check
        alive = uptime > 0  # If we can calculate uptime, we're alive
        
        return {
            "alive": alive,
            "timestamp": timestamp,
            "uptime_seconds": round(uptime, 2)
        }

# Global health checker instance
_health_checker = None

def get_health_checker() -> HealthChecker:
    """Get global health checker instance."""
    global _health_checker
    if _health_checker is None:
        _health_checker = HealthChecker()
    return _health_checker

# Convenience functions for different health checks
async def get_health() -> Dict[str, Any]:
    """Get comprehensive health status."""
    checker = get_health_checker()
    health = await checker.get_comprehensive_health()
    
    # Convert to dictionary for JSON serialization
    result = {
        "status": health.status,
        "timestamp": health.timestamp,
        "version": health.version,
        "uptime_seconds": health.uptime_seconds,
        "summary": health.summary,
        "system_metrics": health.system_metrics,
        "services": []
    }
    
    for service in health.services:
        service_dict = {
            "name": service.name,
            "url": service.url,
            "critical": service.critical,
            "status": asdict(service.status)
        }
        result["services"].append(service_dict)
    
    return result

async def get_health_simple() -> Dict[str, Any]:
    """Get simple health status."""
    checker = get_health_checker()
    return await checker.get_simple_health()

def get_readiness() -> Dict[str, Any]:
    """Get readiness status."""
    checker = get_health_checker()
    return checker.get_readiness_status()

def get_liveness() -> Dict[str, Any]:
    """Get liveness status."""
    checker = get_health_checker()
    return checker.get_liveness_status()

if __name__ == "__main__":
    # Test the health checker
    import asyncio
    
    async def test_health_checks():
        print("Testing health checks...")
        
        print("\n1. Simple Health:")
        simple_health = await get_health_simple()
        print(json.dumps(simple_health, indent=2))
        
        print("\n2. Readiness:")
        readiness = get_readiness()
        print(json.dumps(readiness, indent=2))
        
        print("\n3. Liveness:")
        liveness = get_liveness()
        print(json.dumps(liveness, indent=2))
        
        print("\n4. Comprehensive Health:")
        comprehensive = await get_health()
        print(json.dumps(comprehensive, indent=2))
    
    asyncio.run(test_health_checks())