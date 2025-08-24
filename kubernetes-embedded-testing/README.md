# ket - Kubernetes Embedded Testing

`ket` is a lightweight testing framework that runs your tests inside Kubernetes pods with traffic interception capabilities.

## Quick Start

```bash
# Build the binary
make build

# Run tests against a pod
./bin/ket -target-pod my-app -test-command "npm test" -proc "npm start"
```

## Usage

### Launch Mode (Default)
```bash
ket -target-pod <pod-name> -test-command <test-command> -proc <process-to-test>
```

**Required Flags:**
- `-target-pod`: Name of the pod to test against
- `-test-command`: Command to run tests (e.g., "npm test", "go test ./...")
- `-proc`: Process to test (e.g., "npm start", "go run main.go")

**Optional Flags:**
- `-target-namespace`: Target namespace (default: "default")
- `-project-root`: Project directory (default: ".")
- `-image`: Runner image (default: "node:18-alpine")
- `-debug`: Enable debug logging
- `-keep-namespace`: Keep test namespace after completion

### Examples

```bash
# Test a Node.js app
ket -target-pod express-server -test-command "npm test" -proc "npm start"

# Test a Go app
ket -target-pod go-app -test-command "go test ./..." -proc "go run main.go"

# Keep namespace for debugging
ket -target-pod my-app -test-command "npm test" -proc "npm start" -keep-namespace
```

## How It Works

1. **Creates isolated namespace** with UUID for test isolation
2. **Mounts your source code** into a test pod
3. **Intercepts traffic** from the target pod using mirrord
4. **Runs your tests** in the Kubernetes environment
5. **Streams results** to stdout in real-time
6. **Cleans up** automatically (unless `-keep-namespace` is used)

## Architecture Support

- **x86_64 Linux**: Full mirrord support with traffic interception
- **ARM64 Linux**: Graceful fallback without traffic interception
- **Multi-platform**: Built for both architectures

## Building

```bash
# Build binary
make build

# Build Docker image
make docker-build

# Build for all platforms
make build-all
```

## Requirements

- Go 1.24+
- Kubernetes cluster (Kind, Minikube, etc.)
- kubectl configured
