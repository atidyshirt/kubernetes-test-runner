# ket - Kubernetes Embedded Testing

A Go-based tool for running integration tests in Kubernetes with traffic interception capabilities.

## Overview

`ket` creates isolated test environments in Kubernetes, mounts your source code, and executes tests with optional traffic interception via mirrord.

## Features

- **Isolated Testing**: Unique namespace per test run
- **HostPath Mounting**: Direct source code access (Kind-optimized)
- **Mirrord Integration**: Optional traffic interception
- **Automatic Cleanup**: Resources cleaned up after completion
- **Multi-language Support**: Node.js, Go, Python, etc.

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

### Basic Test Execution

```bash
ket launch \
  --target-pod my-app \
  --target-namespace default \
  --test-command "npm test" \
  --image node:18-alpine
```

### With Traffic Interception

```bash
ket launch \
  --target-pod my-app \
  --target-namespace default \
  --test-command "npm test" \
  --mirrord-process "node server.js" \
  --steal \
  --image node:18-alpine
```

## Command Reference

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--debug` | Enable debug logging | `false` |
| `--image` | Runner image | `node:18-alpine` |
| `--kind-workspace-path` | Kind workspace path | `/workspace` |
| `--project-root` | Project root path | `.` |

### Launch Flags

| Flag | Description | Default | Required |
|------|-------------|---------|----------|
| `--target-pod` | Target pod to test | - | ✅ |
| `--target-namespace` | Target namespace | `default` | ✅ |
| `--test-command` | Test command | - | ✅ |
| `--mirrord-process` | Process for mirrord | - | ❌ |
| `--steal` | Enable steal mode | `false` | ❌ |
| `--keep-namespace` | Keep test namespace | `false` | ❌ |
| `--backoff-limit` | Job backoff limit | `1` | ❌ |
| `--active-deadline-seconds` | Job deadline | `1800` | ❌ |

## Architecture

```
Local → ket launch → Isolated Namespace → Test Job → Results → Cleanup
```

1. **Namespace Creation**: Unique `ket-<uuid>` namespace
2. **Job Deployment**: Test runner with source code mounted
3. **Mirrord Setup**: Optional traffic interception
4. **Test Execution**: Your test command runs
5. **Result Collection**: Test output streamed to stdout
6. **Cleanup**: Namespace and resources removed

## Development

### Project Structure

```
pkg/
├── config/     # Configuration and CLI flags
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
