"""Configuration management for Ollama Monitoring Stack."""

import os
import yaml
import logging
from pathlib import Path
from typing import Dict, Any, Optional
from dataclasses import dataclass, field

# Setup logging
logger = logging.getLogger(__name__)

@dataclass
class ServerConfig:
    """Server configuration settings."""
    ollama_url: str = "http://localhost:11434"
    proxy_port: int = 11435
    proxy_host: str = "localhost"
    metrics_port: int = 8001
    metrics_host: str = "localhost"
    dashboard_port: int = 3001
    dashboard_host: str = "localhost"
    prometheus_port: int = 9090
    prometheus_host: str = "localhost"

@dataclass
class ModelConfig:
    """Model configuration settings."""
    default_model: str = "phi3:mini"
    available_models: list = field(default_factory=lambda: ["phi3:mini", "llama2", "codellama"])

@dataclass
class MonitoringConfig:
    """Monitoring configuration settings."""
    metrics_interval: int = 5
    request_timeout: int = 30
    max_concurrent_requests: int = 10
    max_queue_size: int = 50
    rate_limit_per_minute: int = 60
    
    # AI status generation settings
    ai_status_enabled: bool = True
    ai_status_generation_interval: int = 15
    ai_status_timeout: int = 10
    high_load_active_requests: int = 5
    high_load_queue_size: int = 10

@dataclass
class LoggingConfig:
    """Logging configuration settings."""
    level: str = "INFO"
    format: str = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
    proxy_log: str = "proxy.log"
    dashboard_log: str = "dashboard.log"
    ollama_log: str = "ollama.log"
    prometheus_log: str = "prometheus.log"
    max_size: str = "10MB"
    backup_count: int = 5

@dataclass
class HealthCheckConfig:
    """Health check configuration settings."""
    enabled: bool = True
    interval: int = 30
    timeout: int = 5

class ConfigManager:
    """Configuration manager for the monitoring stack."""
    
    def __init__(self, config_path: Optional[str] = None):
        """Initialize configuration manager.
        
        Args:
            config_path: Path to configuration file. Defaults to config.yml
        """
        self.config_path = config_path or "config.yml"
        self._config = None
        self._server = None
        self._models = None
        self._monitoring = None
        self._logging = None
        self._health_check = None
        
        # Load configuration
        self.load_config()
    
    def load_config(self) -> Dict[str, Any]:
        """Load configuration from YAML file and environment variables."""
        config = {}
        
        # Load from YAML file if it exists
        if os.path.exists(self.config_path):
            try:
                with open(self.config_path, 'r') as f:
                    config = yaml.safe_load(f) or {}
                logger.info(f"Loaded configuration from {self.config_path}")
            except Exception as e:
                logger.warning(f"Failed to load config file {self.config_path}: {e}")
                config = {}
        else:
            logger.info(f"Config file {self.config_path} not found, using defaults")
        
        # Override with environment variables
        env_overrides = self._get_env_overrides()
        config = self._merge_config(config, env_overrides)
        
        self._config = config
        self._initialize_config_objects()
        
        return config
    
    def _get_env_overrides(self) -> Dict[str, Any]:
        """Get configuration overrides from environment variables."""
        overrides = {}
        
        # Server configuration from environment
        if os.getenv('OLLAMA_URL'):
            overrides.setdefault('server', {})['ollama_url'] = os.getenv('OLLAMA_URL')
        if os.getenv('PROXY_PORT'):
            overrides.setdefault('server', {})['proxy_port'] = int(os.getenv('PROXY_PORT'))
        if os.getenv('METRICS_PORT'):
            overrides.setdefault('server', {})['metrics_port'] = int(os.getenv('METRICS_PORT'))
        if os.getenv('DASHBOARD_PORT'):
            overrides.setdefault('server', {})['dashboard_port'] = int(os.getenv('DASHBOARD_PORT'))
        
        # Model configuration from environment
        if os.getenv('DEFAULT_MODEL'):
            overrides.setdefault('models', {})['default_model'] = os.getenv('DEFAULT_MODEL')
        
        # Monitoring configuration from environment
        if os.getenv('REQUEST_TIMEOUT'):
            overrides.setdefault('monitoring', {})['request_timeout'] = int(os.getenv('REQUEST_TIMEOUT'))
        if os.getenv('MAX_CONCURRENT_REQUESTS'):
            overrides.setdefault('monitoring', {})['max_concurrent_requests'] = int(os.getenv('MAX_CONCURRENT_REQUESTS'))
        
        # Logging configuration from environment
        if os.getenv('LOG_LEVEL'):
            overrides.setdefault('logging', {})['level'] = os.getenv('LOG_LEVEL')
        
        return overrides
    
    def _merge_config(self, base: Dict[str, Any], overrides: Dict[str, Any]) -> Dict[str, Any]:
        """Merge configuration dictionaries recursively."""
        result = base.copy()
        
        for key, value in overrides.items():
            if key in result and isinstance(result[key], dict) and isinstance(value, dict):
                result[key] = self._merge_config(result[key], value)
            else:
                result[key] = value
        
        return result
    
    def _initialize_config_objects(self):
        """Initialize typed configuration objects."""
        # Create default instances to get default values
        default_server = ServerConfig()
        default_model = ModelConfig()
        default_monitoring = MonitoringConfig()
        default_logging = LoggingConfig()
        default_health_check = HealthCheckConfig()
        
        # Server configuration
        server_config = self._config.get('server', {})
        self._server = ServerConfig(
            ollama_url=server_config.get('ollama_url', default_server.ollama_url),
            proxy_port=server_config.get('proxy_port', default_server.proxy_port),
            proxy_host=server_config.get('proxy_host', default_server.proxy_host),
            metrics_port=server_config.get('metrics_port', default_server.metrics_port),
            metrics_host=server_config.get('metrics_host', default_server.metrics_host),
            dashboard_port=server_config.get('dashboard_port', default_server.dashboard_port),
            dashboard_host=server_config.get('dashboard_host', default_server.dashboard_host),
            prometheus_port=server_config.get('prometheus_port', default_server.prometheus_port),
            prometheus_host=server_config.get('prometheus_host', default_server.prometheus_host)
        )
        
        # Model configuration
        models_config = self._config.get('models', {})
        self._models = ModelConfig(
            default_model=models_config.get('default_model', default_model.default_model),
            available_models=models_config.get('available_models', default_model.available_models)
        )
        
        # Monitoring configuration
        monitoring_config = self._config.get('monitoring', {})
        ai_status_config = monitoring_config.get('ai_status', {})
        high_load_config = ai_status_config.get('high_load_threshold', {})
        
        self._monitoring = MonitoringConfig(
            metrics_interval=monitoring_config.get('metrics_interval', default_monitoring.metrics_interval),
            request_timeout=monitoring_config.get('request_timeout', default_monitoring.request_timeout),
            max_concurrent_requests=monitoring_config.get('max_concurrent_requests', default_monitoring.max_concurrent_requests),
            max_queue_size=monitoring_config.get('max_queue_size', default_monitoring.max_queue_size),
            rate_limit_per_minute=monitoring_config.get('rate_limit_per_minute', default_monitoring.rate_limit_per_minute),
            ai_status_enabled=ai_status_config.get('enabled', default_monitoring.ai_status_enabled),
            ai_status_generation_interval=ai_status_config.get('generation_interval', default_monitoring.ai_status_generation_interval),
            ai_status_timeout=ai_status_config.get('timeout', default_monitoring.ai_status_timeout),
            high_load_active_requests=high_load_config.get('active_requests', default_monitoring.high_load_active_requests),
            high_load_queue_size=high_load_config.get('queue_size', default_monitoring.high_load_queue_size)
        )
        
        # Logging configuration
        logging_config = self._config.get('logging', {})
        files_config = logging_config.get('files', {})
        
        self._logging = LoggingConfig(
            level=logging_config.get('level', default_logging.level),
            format=logging_config.get('format', default_logging.format),
            proxy_log=files_config.get('proxy', default_logging.proxy_log),
            dashboard_log=files_config.get('dashboard', default_logging.dashboard_log),
            ollama_log=files_config.get('ollama', default_logging.ollama_log),
            prometheus_log=files_config.get('prometheus', default_logging.prometheus_log),
            max_size=logging_config.get('max_size', default_logging.max_size),
            backup_count=logging_config.get('backup_count', default_logging.backup_count)
        )
        
        # Health check configuration
        health_config = self._config.get('health_check', {})
        self._health_check = HealthCheckConfig(
            enabled=health_config.get('enabled', default_health_check.enabled),
            interval=health_config.get('interval', default_health_check.interval),
            timeout=health_config.get('timeout', default_health_check.timeout)
        )
    
    @property
    def server(self) -> ServerConfig:
        """Get server configuration."""
        return self._server
    
    @property
    def models(self) -> ModelConfig:
        """Get model configuration."""
        return self._models
    
    @property
    def monitoring(self) -> MonitoringConfig:
        """Get monitoring configuration."""
        return self._monitoring
    
    @property
    def logging(self) -> LoggingConfig:
        """Get logging configuration."""
        return self._logging
    
    @property
    def health_check(self) -> HealthCheckConfig:
        """Get health check configuration."""
        return self._health_check
    
    def get(self, key: str, default: Any = None) -> Any:
        """Get a configuration value by key."""
        return self._config.get(key, default)
    
    def get_nested(self, path: str, default: Any = None) -> Any:
        """Get a nested configuration value by dot-separated path."""
        keys = path.split('.')
        value = self._config
        
        for key in keys:
            if isinstance(value, dict) and key in value:
                value = value[key]
            else:
                return default
        
        return value
    
    def save_config(self, path: Optional[str] = None):
        """Save current configuration to YAML file."""
        output_path = path or self.config_path
        
        try:
            with open(output_path, 'w') as f:
                yaml.dump(self._config, f, default_flow_style=False, indent=2)
            logger.info(f"Configuration saved to {output_path}")
        except Exception as e:
            logger.error(f"Failed to save configuration to {output_path}: {e}")
            raise
    
    def reload_config(self):
        """Reload configuration from file and environment."""
        self.load_config()
        logger.info("Configuration reloaded")

# Global configuration instance
_config_manager = None

def get_config(config_path: Optional[str] = None) -> ConfigManager:
    """Get the global configuration manager instance."""
    global _config_manager
    
    if _config_manager is None or config_path:
        _config_manager = ConfigManager(config_path)
    
    return _config_manager

# Convenience functions for common configuration access
def get_server_config() -> ServerConfig:
    """Get server configuration."""
    return get_config().server

def get_model_config() -> ModelConfig:
    """Get model configuration."""
    return get_config().models

def get_monitoring_config() -> MonitoringConfig:
    """Get monitoring configuration."""
    return get_config().monitoring

def get_logging_config() -> LoggingConfig:
    """Get logging configuration."""
    return get_config().logging

def get_health_check_config() -> HealthCheckConfig:
    """Get health check configuration."""
    return get_config().health_check

if __name__ == "__main__":
    # Example usage and testing
    config = get_config()
    
    print("Server Configuration:")
    print(f"  Ollama URL: {config.server.ollama_url}")
    print(f"  Proxy Port: {config.server.proxy_port}")
    print(f"  Metrics Port: {config.server.metrics_port}")
    print(f"  Dashboard Port: {config.server.dashboard_port}")
    
    print("\nModel Configuration:")
    print(f"  Default Model: {config.models.default_model}")
    print(f"  Available Models: {config.models.available_models}")
    
    print("\nMonitoring Configuration:")
    print(f"  Request Timeout: {config.monitoring.request_timeout}s")
    print(f"  Max Concurrent Requests: {config.monitoring.max_concurrent_requests}")
    print(f"  AI Status Generation: {config.monitoring.ai_status_enabled}")