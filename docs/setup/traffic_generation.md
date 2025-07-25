# Ollama Traffic Generator

This system generates continuous prompt traffic to Ollama for monitoring model performance and collecting metrics.

## Overview

The traffic generator uses 1000 curated questions across 10 categories:
- **General Knowledge** (100 questions) - Basic facts and trivia
- **Science** (100 questions) - Physics, chemistry, biology, earth science
- **Technology** (100 questions) - Computers, software, internet, modern tech
- **History** (100 questions) - World history, wars, civilizations
- **Geography** (100 questions) - Countries, capitals, physical geography
- **Sports** (100 questions) - Various sports, rules, athletes
- **Entertainment** (100 questions) - Movies, music, TV, celebrities
- **Literature** (100 questions) - Books, authors, characters
- **Philosophy** (100 questions) - Philosophical concepts, thinkers
- **Food** (100 questions) - Cuisine, cooking, ingredients

## Files Structure

```
├── traffic_generator.py          # Main traffic generator script
├── questions/                    # Question category files
│   ├── general_knowledge.json
│   ├── science.json
│   ├── technology.json
│   ├── history.json
│   ├── geography.json
│   ├── sports.json
│   ├── entertainment.json
│   ├── literature.json
│   ├── philosophy.json
│   └── food.json
├── requirements.txt              # Python dependencies
└── README_traffic_generator.md   # This file
```

## Installation

1. **Install dependencies for traffic generator:**
   ```bash
   pip install -r requirements_traffic.txt
   # OR just: pip install aiohttp
   ```

   **If you also want to run the LlamaIndex app.py:**
   ```bash
   pip install -r requirements_app.txt
   ```

2. **Make sure Ollama is running:**
   ```bash
   ollama serve
   ```

3. **Ensure you have the model installed:**
   ```bash
   ollama pull phi-3:mini
   ```

## Usage

### Basic Usage
```bash
python traffic_generator.py
```

### Advanced Options
```bash
# Use a different model
python traffic_generator.py --model llama2

# Limit to 50 questions
python traffic_generator.py --max 50

# Change delay between requests (in seconds)
python traffic_generator.py --delay 2.0

# Use different Ollama URL
python traffic_generator.py --url http://localhost:11434
```

### Command Line Arguments
- `--model`: Ollama model to use (default: phi-3:mini)
- `--url`: Ollama API URL (default: http://localhost:11434)
- `--max`: Maximum number of questions to ask (default: unlimited)
- `--delay`: Delay between requests in seconds (default: 1.0)

## Example Output

```
2024-01-15 10:30:00 - INFO - 🚀 Starting traffic generator for model: phi-3:mini
2024-01-15 10:30:00 - INFO - 📚 Total questions loaded: 1000
2024-01-15 10:30:00 - INFO - 📁 Categories: General Knowledge, Science, Technology, History, Geography, Sports, Entertainment, Literature, Philosophy, Food
2024-01-15 10:30:00 - INFO - ✅ Ollama is running and accessible
2024-01-15 10:30:01 - INFO - ✅ Q1: What is the capital of France?... (Latency: 0.85s)
2024-01-15 10:30:02 - INFO - ✅ Q2: What is Newton's first law of motion?... (Latency: 1.23s)
...
```

## Monitoring Integration

The traffic generator works with your existing monitoring setup:

1. **Prometheus Metrics**: Start your app.py to expose metrics on port 8000
2. **Prometheus Server**: Use run_prometheus.sh to start monitoring
3. **Traffic Generation**: Run this script to generate load

## Features

- **Serialized Processing**: Questions are asked one at a time, waiting for responses
- **Random Selection**: Questions are randomly selected from the 1000-question pool
- **Health Checking**: Verifies Ollama is accessible before starting
- **Statistics Tracking**: Tracks success rates, latencies, and error counts
- **Graceful Shutdown**: Handles Ctrl+C interrupts cleanly
- **Logging**: Comprehensive logging to both console and file
- **Error Handling**: Robust error handling with helpful messages

## Troubleshooting

### "Cannot connect to Ollama"
- Ensure Ollama is running: `ollama serve`
- Check if the model is available: `ollama list`
- Verify the URL is correct (default: http://localhost:11434)

### "No questions loaded"
- Ensure the `questions/` directory exists
- Check that JSON files are valid
- Verify file permissions

### High error rates
- Check if the model is compatible with your system
- Increase the delay between requests
- Monitor system resources (CPU, memory)

## Statistics

The generator tracks and displays:
- Total requests sent
- Successful responses
- Failed requests
- Success rate percentage
- Average response latency
- Real-time progress updates

## Integration with Monitoring

This traffic generator is designed to work with:
- **app.py**: LlamaIndex application with Prometheus metrics
- **Prometheus**: Time-series metrics collection
- **Grafana**: Visualization dashboards (if configured)

The generated traffic will appear in your monitoring dashboards, allowing you to observe:
- Query rates
- Response times
- Error rates
- System performance under load