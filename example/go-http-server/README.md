# Go HTTP Server Example

This example demonstrates how to test a Go HTTP server using the `ket` (Kubernetes Embedded Testing) framework.

## Prerequisites

- Docker
- kind (Kubernetes in Docker)
- kubectl

## Cluster Setup

### 1. Create kind cluster with hostPath mount

```bash
# Create kind cluster with current directory mounted
kind create cluster --config kind-config.yaml
```

The `kind-config.yaml` mounts the current directory to `/workspace` inside the kind container.

### 2. Build and deploy the Go HTTP server

```bash
# Build the Go server image
docker build -t go-http-server:latest .

# Load into kind cluster
kind load docker-image go-http-server:latest

# Deploy the server
kubectl apply -f manifest/go-server-pod.yaml

# Wait for server to be ready
kubectl wait --for=condition=ready pod/go-http-server --timeout=60s
```

## Running Tests with ket

```bash
# From the go-http-server directory
ket launch \
  --target-pod go-http-server \
  --target-namespace default \
  --test-command "go test -v ./test/..." \
  --proc "echo 'dummy'" \
  --project-root . \
  --image golang:1.24-alpine \
  --kind-workspace-path /workspace
```

## Server Endpoints

- `GET /` - Returns a welcome message
- `GET /health` - Health check endpoint
- `POST /api/echo` - Echo endpoint that returns the request body

## Test Structure

The tests use Go's standard testing package and `httptest` for mocking HTTP requests:

- Unit tests for individual endpoints
- Integration tests for multiple endpoints
- Error handling tests (method not allowed, invalid JSON)

## Cleanup

```bash
# Delete the server
kubectl delete -f manifest/go-server-pod.yaml

# Delete the kind cluster
kind delete cluster
```

## Architecture

The test runner uses `mirrord --steal` to intercept traffic from the running Go server pod, allowing tests to run against the actual service while maintaining isolation.
