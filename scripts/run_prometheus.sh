#!/bin/bash

# This script automates the configuration and launch of Prometheus via Podman
# for the LlamaStack monitoring project. It should be run from your
# project directory (e.g., 'llamastack-prometheus').

# --- Step 1: Check for and Create the Prometheus Configuration File ---
CONFIG_FILE="prometheus.yml"

if [ -f "$CONFIG_FILE" ]; then
    echo "--- Found existing '$CONFIG_FILE', skipping creation. ---"
else
    echo "--- Creating '$CONFIG_FILE' configuration file ---"
    # Use a heredoc (cat << EOL ... EOL) to write the multi-line configuration
    # into a file named prometheus.yml in the current directory.
    cat > $CONFIG_FILE << EOL
global:
  scrape_interval: 15s # How often to scrape metrics

scrape_configs:
  - job_name: 'llamastack_app'
    # This job defines how to find and scrape metrics from your Python app.
    static_configs:
      - targets: ['host.containers.internal:8000']
        # 'host.containers.internal' is a special DNS name for Podman that allows the
        # container to connect to a service running on the host machine (your Mac).
        # It is the equivalent of 'host.docker.internal' for Docker.
EOL
    echo "$CONFIG_FILE created successfully."
fi

echo ""
echo "--- Step 2: Launching Prometheus container with Podman ---"
echo "Make sure the Podman machine is running."
echo ""

# --- Step 3: Run the Prometheus Container using Podman in Background ---
# This uses 'podman run' instead of 'docker run'. The flags are compatible.
# --rm: Automatically removes the container when it exits.
# --name prometheus: Assigns a name to the container for easy reference.
# -p 9090:9090: Maps port 9090 on your Mac to port 9090 inside the container.
# -v "$(pwd)/prometheus.yml":...:Z: Mounts your local prometheus.yml file
#    into the expected configuration path inside the container. The ':Z' handles SELinux permissions.
# -d: Runs the container in detached mode (background).
# docker.io/prom/prometheus: The full image path for the official Prometheus image.
podman run --rm --name prometheus \
-p 9090:9090 \
-v "$(pwd)/prometheus.yml":/etc/prometheus/prometheus.yml:Z \
-d \
docker.io/prom/prometheus

# Wait a moment for the container to start
sleep 2

# Check if the container is running and print the URL
if podman ps --format "table {{.Names}}" | grep -q prometheus; then
    echo "✅ Prometheus is running successfully!"
    echo "🌐 Prometheus URL: http://localhost:9090"
    echo ""
    echo "To stop the container, run: podman stop prometheus"
    echo "To view logs, run: podman logs prometheus"
else
    echo "❌ Failed to start Prometheus container"
    echo "Check if Podman is running and try again"
fi

