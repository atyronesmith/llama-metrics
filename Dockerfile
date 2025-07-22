# Ollama Monitoring Stack Dockerfile
FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    procps \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements and install Python dependencies
COPY requirements_all.txt .
RUN pip install --no-cache-dir -r requirements_all.txt

# Copy application files
COPY *.py ./
COPY config.yml ./
COPY VERSION ./
COPY docs/ ./docs/
COPY templates/ ./templates/

# Create non-root user
RUN useradd -m -u 1000 monitor && \
    chown -R monitor:monitor /app
USER monitor

# Expose ports
EXPOSE 11435 8001 3001

# Health check using our new endpoints
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8001/health/simple || exit 1

# Environment variables
ENV PYTHONUNBUFFERED=1
ENV LOG_LEVEL=INFO

# Default command - run the monitoring proxy
CMD ["python", "ollama_monitoring_proxy_fixed.py"]