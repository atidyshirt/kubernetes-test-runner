# ket - Kubernetes Embedded Testing

A Go-based tool for deploying and running tests in Kubernetes environments.

## Overview

`ket` creates isolated test environments in Kubernetes, mounts your source code, and executes tests in an isolated namespace.

## Features

- **Isolated Testing**: Unique namespace per test run
- **HostPath Mounting**: Direct source code access (Kind-optimized)
- **Automatic Cleanup**: Resources cleaned up after completion
- **Multi-language Support**: Node.js, Go, Python, etc.
- **Config File Support**: YAML/JSON configuration files

## Installation

```bash
# Build from source
make build

# Install dependencies
go mod download

# Run tests
make test
```

## Usage

### Using Command Line Options

```bash
ket launch \
  --test-command "npm test" \
  --image node:18-alpine
```

### Using Configuration File

Create a `ket-config.yaml` file:

```yaml
mode: launch
projectRoot: .
image: node:18-alpine
debug: false
testCommand: npm test
keepNamespace: false
backoffLimit: 1
activeDeadlineS: 1800
kindWorkspacePath: /workspace
```

Then run:

```bash
ket launch --config ket-config.yaml
```

### Command Line Options Override Config File

You can combine config files with command line options. Command line options take precedence:

```bash
ket launch --config ket-config.yaml --debug --keep-namespace
```

## Command Reference

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config, -c` | Path to config file | - |
| `--debug, -v` | Enable debug logging | `false` |
| `--image, -i` | Runner image | `node:18-alpine` |
| `--kind-workspace-path, -w` | Kind workspace path | `/workspace` |
| `--project-root, -r` | Project root path | `.` |

### Launch Flags

| Flag | Description | Default | Required |
|------|-------------|---------|----------|
| `--test-command, -t` | Test command to execute | - | ✅ |
| `--keep-namespace, -k` | Keep test namespace | `false` | ❌ |
| `--backoff-limit, -b` | Job backoff limit | `1` | ❌ |
| `--active-deadline-seconds, -d` | Job deadline in seconds | `1800` | ❌ |

## Architecture

```
Local → ket launch → Isolated Namespace → Test Job → Results → Cleanup
```

1. **Namespace Creation**: Unique `ket-<uuid>` namespace
2. **Job Deployment**: Test runner with source code mounted
3. **Test Execution**: Your test command runs
4. **Result Collection**: Test output streamed to stdout
5. **Cleanup**: Namespace and resources removed

## Development

### Project Structure

```
pkg/
├── config/     # Configuration and file loading
├── kube/       # Kubernetes operations
├── launcher/   # Job launch orchestration
└── logger/     # Structured logging

cmd/
└── testrunner/ # CLI entry point
```

### Building

```bash
make build      # Build binary
make test       # Run tests
make lint       # Lint code
make clean      # Clean artifacts
```

### Testing

```bash
# Run all tests
go test -v ./...

# Run specific package
go test -v ./pkg/kube

# Run with coverage
go test -v -coverprofile=coverage.out ./...
```

## Examples

See the `example/` directory for working examples:
- **Node.js**: Express server with Mocha tests
- **Go**: HTTP server with Go tests

## Requirements

- Go 1.24+
- Kubernetes cluster (Kind recommended)
- kubectl configured
- Docker (for building)

## Docker Image

The project includes a simplified Dockerfile that creates a test runner image with essential tools:

- **Base**: Ubuntu 24.04
- **kubectl**: Copied from official `bitnami/kubectl` image
- **Node.js**: Latest LTS version (22.x)
- **mirrord**: Installed using official MetalBear installation script

To build the image:

```bash
docker build -t ket-test-runner .
```
