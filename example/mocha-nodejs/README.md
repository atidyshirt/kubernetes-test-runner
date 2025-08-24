# Node.js Express Server Example with ket

This example demonstrates how to use `ket` (Kubernetes Embedded Testing) to run integration tests against a Node.js Express server.

## Overview

- **Express Server**: Simple REST API with user CRUD operations
- **Tests**: Mocha-based integration tests that run in Kubernetes
- **ket Integration**: Uses `ket launch` to run tests with traffic interception

## Quick Start

### 1. Setup the Express Server

```bash
# Build and deploy the Express server
./setup.sh
```

This will:
- Build the Docker image
- Load it into your Kind cluster
- Deploy the server and service

### 2. Run Tests with ket

```bash
# Run tests with traffic interception
ket launch \
  --target-pod express-server \
  --target-namespace default \
  --test-command "npm test" \
  --proc "npm start" \
  --image node:18-alpine \
  --kind-workspace-path /workspace

# Or run tests without traffic interception
ket launch \
  --target-pod express-server \
  --target-namespace default \
  --test-command "npm test" \
  --image node:18-alpine \
  --kind-workspace-path /workspace
```

## Architecture

- **Express Server**: Runs as a normal Kubernetes deployment
- **ket**: Creates an isolated test namespace and runs tests
- **Traffic Interception**: Uses mirrord to intercept traffic when `--proc` is specified
- **Test Execution**: Tests run in the same cluster with access to service DNS

## Test Structure

- **Unit Tests**: `test/unit/` - Test individual functions
- **Integration Tests**: `test/integration/` - Test API endpoints
- **ket Tests**: Verify ket configuration and integration

## API Endpoints

- `GET /health` - Health check
- `GET /` - API information
- `GET /api/users` - List users
- `POST /api/users` - Create user
- `GET /api/users/:id` - Get user by ID
- `PUT /api/users/:id` - Update user
- `DELETE /api/users/:id` - Delete user

## Requirements

- Kind cluster running
- ket binary built and available
- Docker for building images
