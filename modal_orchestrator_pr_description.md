# Add Modal Orchestrator Integration to ZenML

## 🚀 Overview

This PR introduces a new **Modal Orchestrator** integration for ZenML, enabling users to run complete ML pipelines on Modal's serverless cloud infrastructure with optimized performance and cost efficiency.

## 🎯 What This PR Does

### Core Features
- **🔧 Modal Orchestrator**: New orchestrator flavor that executes entire ZenML pipelines on Modal's cloud platform
- **⚡ Optimized Execution Modes**: Two execution modes for different use cases:
  - `pipeline` (default): Runs entire pipeline in single Modal function for maximum speed
  - `per_step`: Runs each step separately for granular control and debugging
- **🏗️ Persistent Apps**: Implements warm container reuse with dynamic sandboxes for faster execution
- **💾 Resource Configuration**: Full support for GPU, CPU, memory settings with intelligent defaults
- **🔐 Authentication**: Modal token support with fallback to default Modal auth

### Key Benefits
- **Performance**: Running entire pipelines in single containers eliminates inter-step overhead
- **Cost Efficiency**: Fewer container spawns = lower costs on Modal's platform
- **Simplicity**: Clean API with just two execution modes for distinct use cases
- **ZenML Native**: Leverages ZenML's PipelineEntrypointConfiguration for optimal integration

## 📁 File Structure

```
src/zenml/integrations/modal/
├── orchestrators/                        # New directory
│   ├── __init__.py
│   └── modal_orchestrator.py            # Main orchestrator implementation
├── flavors/
│   ├── modal_orchestrator_flavor.py     # New flavor definition
│   └── __init__.py                       # Updated exports
├── utils.py                              # Shared utilities for Modal integration
└── __init__.py                           # Updated with orchestrator registration
```

## 🛠️ Implementation Details

### Execution Architecture
- **Persistent Pipeline Apps**: Each pipeline gets its own persistent Modal app for optimal performance
- **Dynamic Sandboxes**: Each run creates fresh sandboxes for complete isolation
- **Built-in Output Streaming**: Modal automatically handles log streaming and output capture
- **Maximum Flexibility**: Sandboxes can execute arbitrary commands with better isolation

### Resource Management
- **Smart Resource Configuration**: Automatic fallbacks and intelligent defaults
- **GPU Support**: Full GPU type specification (T4, A100, H100) with count configuration
- **Memory & CPU**: Configurable CPU cores and memory allocation
- **Cost Optimization**: Persistent apps with warm containers to minimize cold starts

### Authentication & Security
- **Multiple Auth Methods**: Modal API tokens or CLI authentication
- **Environment Support**: Different Modal environments (dev, staging, production)
- **Workspace Management**: Multi-workspace support for team collaboration

## 📖 Usage Examples

### Basic Usage
```python
from zenml import pipeline
from zenml.integrations.modal.flavors import ModalOrchestratorSettings

# Configure Modal orchestrator
@pipeline(
    settings={
        "orchestrator": ModalOrchestratorSettings(
            execution_mode="pipeline",  # Run entire pipeline in one function
            gpu="A100",
            region="us-east-1"
        )
    }
)
def my_pipeline():
    # Your pipeline steps here
    pass
```

### Advanced Resource Configuration
```python
from zenml.config import ResourceSettings
from zenml.integrations.modal.flavors import ModalOrchestratorSettings

# Configure Modal-specific settings
modal_settings = ModalOrchestratorSettings(
    gpu="A100",                     # GPU type
    region="us-east-1",             # Preferred region
    cloud="aws",                    # Cloud provider
    modal_environment="production", # Modal environment
    execution_mode="per_step",      # Per-step execution
    max_parallelism=3,              # Max concurrent steps
    timeout=3600,                   # 1 hour timeout
    synchronous=True,               # Wait for completion
)

# Configure hardware resources
resource_settings = ResourceSettings(
    cpu_count=16,                   # Number of CPU cores
    memory="32GB",                  # 32GB RAM
    gpu_count=1                     # Number of GPUs
)

@pipeline(
    settings={
        "orchestrator": modal_settings,
        "resources": resource_settings
    }
)
def my_modal_pipeline():
    # Your pipeline steps here
    pass
```

## 🔧 Technical Improvements

### Shared Utilities
- **`utils.py`**: Consolidated common functionality between orchestrator and step operator
- **GPU Configuration**: Unified GPU type and count handling
- **Authentication**: Centralized Modal client setup
- **Resource Validation**: Consistent resource settings validation
- **Stack Validation**: Shared validation logic for remote components

### Step Operator Updates
- **Unified Resource Handling**: Updated to use shared GPU and resource utilities
- **Consistent Authentication**: Uses same auth setup as orchestrator
- **Improved Error Handling**: Better error messages and validation

## 📚 Documentation

### Comprehensive Guide
- **New Documentation**: Complete guide at `docs/book/component-guide/orchestrators/modal.md`
- **Setup Instructions**: Step-by-step setup and configuration
- **Best Practices**: Performance optimization and cost management tips
- **Troubleshooting**: Common issues and solutions
- **Integration Examples**: Real-world usage patterns

### Updated Documentation
- **README Updates**: Added Modal orchestrator to orchestrator list
- **Step Operator Docs**: Updated with orchestrator comparison and guidance

## 🧪 Testing

### Test Updates
- **Unit Tests**: Updated Modal step operator tests to use shared utilities
- **Integration Tests**: Validates GPU configuration and resource handling
- **Validation Tests**: Ensures proper error handling and validation

## 🔄 Migration & Compatibility

### Breaking Changes
- **None**: This is a new integration that doesn't affect existing functionality
- **Backward Compatible**: All existing Modal step operator functionality preserved

### Dependencies
- **Modal Requirement**: Adds `modal>=0.64.49,<1` to Modal integration requirements
- **No Core Changes**: No new dependencies for core ZenML

## 🏗️ Architecture Benefits

### Why This Approach?

1. **Performance**: Running entire pipelines in single containers eliminates inter-step overhead
2. **Cost Efficiency**: Fewer container spawns = lower costs on Modal's platform
3. **Simplicity**: Clean API with just two execution modes for distinct use cases
4. **ZenML Native**: Leverages ZenML's PipelineEntrypointConfiguration for optimal integration
5. **Flexibility**: Persistent apps with dynamic sandboxes provide optimal balance of performance and isolation

### Design Patterns
- **Follows ZenML Conventions**: Uses standard ZenML orchestrator patterns
- **Consistent with Other Orchestrators**: Similar to GCP Vertex, Kubernetes orchestrators
- **Seamless Integration**: Works with existing ZenML stack architecture

## 📋 Checklist

- [x] **Implementation**: Core orchestrator and flavor implemented
- [x] **Documentation**: Comprehensive documentation added
- [x] **Tests**: Unit and integration tests updated
- [x] **Utilities**: Shared utilities for consistent behavior
- [x] **Examples**: Usage examples and best practices
- [x] **Error Handling**: Robust error handling and validation
- [x] **Authentication**: Multiple authentication methods supported
- [x] **Resource Management**: GPU, CPU, memory configuration
- [x] **Performance**: Optimized execution with persistent apps

## 🎉 Result

This PR adds a production-ready Modal orchestrator that provides:
- **Fastest execution** with pipeline mode
- **Granular control** with per-step mode  
- **Cost optimization** through persistent apps
- **Enterprise features** like multi-environment support
- **Seamless integration** with existing ZenML stacks

The Modal orchestrator follows the same patterns as other ZenML orchestrators and integrates seamlessly with the existing ZenML stack architecture, enabling users to leverage Modal's serverless infrastructure for scalable ML pipeline execution.