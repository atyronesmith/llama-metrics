import time
import asyncio
import os
import sys
import signal
from typing import Optional

# Check for required packages and provide helpful error messages
try:
    from llama_index.core import (
        VectorStoreIndex,
        SimpleDirectoryReader,
        Settings
    )
    from llama_index.core.embeddings import resolve_embed_model
    from llama_index.llms.ollama import Ollama
except ImportError as e:
    print("❌ Missing required packages. Please install them:")
    print("  pip install llama-index llama-index-llms-ollama")
    print(f"  Error: {e}")
    sys.exit(1)

try:
    from prometheus_client import start_http_server, Counter, Histogram, Gauge
except ImportError as e:
    print("❌ Missing prometheus_client. Please install it:")
    print("  pip install prometheus-client")
    print(f"  Error: {e}")
    sys.exit(1)

# --- 1. Prometheus Metrics Definition ---
QUERY_COUNTER = Counter(
    'llama_queries_total',
    'Total number of queries made to the LlamaIndex app'
)
QUERY_LATENCY = Histogram(
    'llama_query_latency_seconds',
    'Latency of queries in seconds'
)
DOCUMENTS_LOADED = Gauge(
    'llama_documents_loaded',
    'Number of documents loaded into the index'
)
ERROR_COUNTER = Counter(
    'llama_errors_total',
    'Total number of errors encountered'
)

# Global variable to track if we should exit
should_exit = False

def signal_handler(signum, frame):
    """Handle shutdown signals gracefully"""
    global should_exit
    print("\n🛑 Received shutdown signal. Cleaning up...")
    should_exit = True

# --- 2. One-Time Setup Function ---
def setup_index():
    """
    Performs the initial setup of the LlamaIndex application,
    including model configuration and document indexing.
    Returns the query engine.
    """
    print("--- Configuring LlamaIndex Settings ---")

    try:
        Settings.llm = Ollama(model="phi3:mini", request_timeout=120.0)
        # Use a simple local embedding model that doesn't require heavy dependencies
        try:
            Settings.embed_model = resolve_embed_model("local:BAAI/bge-small-en-v1.5")
            print("✅ Using HuggingFace embeddings")
        except:
            # Fallback to a simple local embedding
            print("⚠️  Using simple local embeddings (no external dependencies)")
            Settings.embed_model = resolve_embed_model("local")
        print("✅ Model configuration successful")
    except Exception as e:
        print(f"❌ Failed to configure models: {e}")
        print("💡 Make sure Ollama is running and the phi3:mini model is available:")
        print("  ollama pull phi3:mini")
        sys.exit(1)

    print("--- Preparing Data ---")
    if not os.path.exists("data"):
        os.makedirs("data")
        print("📁 Created data directory")

    # Create sample data if it doesn't exist
    data_file = "data/facts.txt"
    if not os.path.exists(data_file):
        with open(data_file, "w") as f:
            f.write("The sky is blue during a clear day. The grass is typically green. Water is essential for life.")
        print("📄 Created sample data file")

    try:
        documents = SimpleDirectoryReader(input_dir="./data").load_data()
        doc_count = len(documents)
        DOCUMENTS_LOADED.set(doc_count)
        print(f"📚 Loaded {doc_count} document(s).")
    except Exception as e:
        print(f"❌ Failed to load documents: {e}")
        sys.exit(1)

    print("--- Creating Index (this may take a moment) ---")
    try:
        index = VectorStoreIndex.from_documents(documents)
        print("✅ Index created successfully")
    except Exception as e:
        print(f"❌ Failed to create index: {e}")
        sys.exit(1)

    print("--- Setup Complete ---")
    return index.as_query_engine()

# --- 3. Interactive Chat Loop ---
async def main():
    """
    The main function that runs the interactive chat loop.
    """
    global should_exit

    # Set up signal handlers for graceful shutdown
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    query_engine = setup_index()

    print("\n--- Interactive Chat Started ---")
    print("Type your query and press Enter. Type 'quit' or 'exit' to end the session.")
    print("Press Ctrl+C to stop the application.")

    while not should_exit:
        try:
            # Prompt the user for input
            query = input("\n> ")

            if query.lower() in ['quit', 'exit']:
                print("👋 Exiting chat. Goodbye!")
                break

            if not query.strip():
                continue

            # --- Run Query and Record Metrics ---
            start_time = time.time()
            try:
                QUERY_COUNTER.inc()

                # Use synchronous query since we're in an async context but the engine is sync
                response = query_engine.query(query)
                print(f"\n🤖 Model: {response}")

            except Exception as e:
                ERROR_COUNTER.inc()
                print(f"❌ Query failed: {e}")
                print("💡 Make sure Ollama is running and the model is available")
            finally:
                latency = time.time() - start_time
                QUERY_LATENCY.observe(latency)
                print(f"⏱️  (Latency: {latency:.2f}s)")

        except KeyboardInterrupt:
            print("\n👋 Exiting chat. Goodbye!")
            break
        except EOFError:
            print("\n👋 End of input. Goodbye!")
            break
        except Exception as e:
            ERROR_COUNTER.inc()
            print(f"❌ An unexpected error occurred: {e}")

# --- 4. Main Execution Block ---
if __name__ == "__main__":
    print("🚀 Starting LlamaStack Prometheus Monitoring App")

    # Start the Prometheus metrics server
    try:
        start_http_server(8000)
        print("📊 Prometheus metrics server started at http://localhost:8000/metrics")
    except Exception as e:
        print(f"❌ Failed to start Prometheus server: {e}")
        sys.exit(1)

    # Run the main interactive loop
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n🛑 Shutting down.")
    finally:
        print("✅ Shutdown complete.")

