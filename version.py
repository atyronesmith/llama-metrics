"""Version information for Ollama Monitoring Stack."""

import os
from pathlib import Path

# Read version from VERSION file
VERSION_FILE = Path(__file__).parent / "VERSION"

def get_version():
    """Get the current version from VERSION file."""
    try:
        with open(VERSION_FILE, 'r') as f:
            return f.read().strip()
    except FileNotFoundError:
        return "unknown"

__version__ = get_version()

# Version components for programmatic access
version_info = tuple(int(x) for x in __version__.split('.') if x.isdigit())

# Build info
BUILD_INFO = {
    "version": __version__,
    "version_info": version_info,
    "project": "Ollama Monitoring Stack",
    "description": "Comprehensive monitoring solution for Ollama AI models",
    "author": "Claude Code Assistant",
    "platform": "Mac M-Series optimized"
}

def print_version():
    """Print version and build information."""
    print(f"{BUILD_INFO['project']} v{BUILD_INFO['version']}")
    print(f"Description: {BUILD_INFO['description']}")
    print(f"Platform: {BUILD_INFO['platform']}")

if __name__ == "__main__":
    print_version()