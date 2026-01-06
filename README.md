# golms

A Local Model Server Interface written in Go for managing and interacting with LLM model servers. This was initially built as a challenge for myself to code a useful tool in Go during a 6-hour flight with limited internet access (Chang Min).

## Overview

`golms` is a CLI tool that provides a unified interface for managing and chatting with local LLM models across different model servers. It currently supports:
- **MLX LM** - Apple Silicon optimized model server
- **Ollama** - Cross-platform model server (WIP)

## Features

- 📋 List available LLMs and model servers
- 🔌 Connect to model servers and chat with LLMs
- 🎯 Automatic model server discovery and management
- 💬 Interactive chat interface with styled TUI
- 🎨 Beautiful terminal UI with color-coded messages and status indicators
- 🚀 Support for multiple model server backends

## Prerequisites

Make sure you have the following installed:
- **Go** (1.25.1 or later)
- **Model Server** (at least one of the supported servers):
  - [MLX LM](https://github.com/ml-explore/mlx-examples) for Apple Silicon
  - [Ollama](https://ollama.ai/) for cross-platform support
- **LLM Models** downloaded in `~/golms/<model_server>/` directories
  - For example: `~/golms/mlx_lm/` or `~/golms/ollama/`

## Installation

### Using go install (Recommended)

```bash
go install github.com/changminbark/golms@latest
```

This will install the latest version to `$GOPATH/bin`. Make sure `$GOPATH/bin` is in your `$PATH`.

### From Source

```bash
# Clone the repository
git clone https://github.com/changminbark/golms.git
cd golms

# Build the binary
make build

# Or install to $GOPATH/bin
make install
```

### Build Modes

#### Default Mode (uses cache, ignores vendor/)
```bash
go build
```
Uses dependencies from `~/go/pkg/mod/`

#### Vendor Mode (uses vendor/)
```bash
go build -mod=vendor
```
Uses dependencies from `./vendor/`

## Usage

### List Available LLMs and Model Servers

```bash
golms list
```

This will show all available LLMs organized by model server and list all installed model servers.

### List Supported Model Servers

```bash
golms servers
```

Shows all model servers that `golms` supports.

### Connect to a Model Server and Chat

```bash
golms connect
```

This will:
1. Prompt you to select a model server
2. Show available LLMs for that server
3. Start the model server (if not already running)
4. Connect you to an interactive chat session

## Project Structure

```
golms/
├── cmd/
│   └── root.go              # CLI commands and handlers
├── pkg/
│   ├── client/              # Client implementations for model servers
│   │   ├── client.go
│   │   ├── mlx_lm.go
│   │   └── ollama.go
│   ├── constants/           # Constants and configurations
│   │   └── model_server.go
│   ├── discovery/           # Model and server discovery
│   │   └── discovery.go
│   ├── server/              # Model server management
│   │   ├── manager.go
│   │   ├── mlx_lm.go
│   │   └── ollama.go
│   ├── ui/                  # Terminal UI styles and formatting
│   │   └── styles.go
│   └── utils/               # Utility functions
│       ├── clean.go
│       └── clean_test.go
├── main.go                  # Application entry point
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Development

### Building

```bash
# Build the binary
make build

# Build with vendor dependencies
make build-vendor
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run go vet
make vet

# Run all checks (fmt, vet, test)
make check
```

### Cleaning

```bash
# Remove binary and clean build cache
make clean
```

## Commands Reference

| Command | Description |
|---------|-------------|
| `golms` | Show usage information |
| `golms list` | List all available LLMs and model servers |
| `golms servers` | List all supported model servers |
| `golms connect` | Connect to a model server and start chatting with an LLM |

## Configuration

Models should be placed in the following directory structure:
```
~/golms/
├── mlx_lm/
│   ├── model-1/
│   └── model-2/
└── ollama/
    ├── model-3/
    └── model-4/
```

Each model directory should contain the necessary model weights and configuration files required by the respective model server.

## License

See LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
