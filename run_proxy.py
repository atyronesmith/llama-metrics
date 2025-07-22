# run_proxy.py
from litellm.proxy.proxy_server import start_proxy_server

# Define the list of models you want the proxy to serve
model_list = [
    {
        "model_name": "local-llama", # A name for your app to call this model
        "litellm_params": {
            "model": "ollama/llama3", # Tell LiteLLM to use the ollama provider
            "api_base": "http://localhost:11434" # The address of your Ollama server
        }
    }
]

# This is the correct function call for your version
start_proxy_server(
    model_list=model_list,
    prometheus=True,
    port=8000
)
