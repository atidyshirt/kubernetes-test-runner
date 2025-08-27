# Node.js + TypeScript Integration Test Example

This example demonstrates how to use ket for running integration tests with Node.js, TypeScript, and MongoDB.

## Structure

- `src/example-http-server.ts` - Express server that writes to MongoDB
- `manifests/` - Kubernetes manifests for MongoDB and HTTP server
- `test/helper/` - Test utilities for managing cluster state and services
- `test/integration/` - Integration tests that verify the full flow

## Setup

1. Install dependencies:
   ```bash
   npm install
   ```

2. Build the TypeScript:
   ```bash
   npm run build
   ```

3. Ensure you have a Kubernetes cluster running (Kind recommended)

4. Build the ket test runner image:
   ```bash
   docker build -t ket-test-runner .
   ```

## Running Tests

### Integration Tests (via ket)
```bash
npm run test:integration
```

This will:
1. Deploy MongoDB and HTTP server to Kubernetes
2. Run mirrord with --steal to intercept traffic
3. Execute the integration tests
4. Clean up all resources

**Note**: Integration tests must be run via ket, not directly with Mocha, as they require a Kubernetes cluster environment.

### Direct Test Execution (for development only)
```bash
# Create test namespace first
npm run cluster:create-ns

# Then run tests (will fail without proper cluster setup)
npm run test:internal:mocha-integration
```

## Complete Workflow Example

```bash
# 1. Setup everything
npm run setup

# 2. Create test namespace
npm run cluster:create-ns

# 3. Run integration tests
npm run test:integration

# 4. Cleanup when done
npm run cleanup
```

## How It Works

1. **TestContainer** manages the test lifecycle
2. **KubectlServiceManager** handles Kubernetes operations
3. **MongoService** provides MongoDB operations using cluster service names
4. **MockServiceManager** runs mirrord to intercept traffic
5. Tests verify HTTP server functionality and data persistence

## Cluster Communication

Since tests run inside the Kubernetes cluster via ket:
- **MongoDB**: Uses `mongodb:27017` (cluster service name)
- **HTTP Server**: Uses `example-http-server:3000` (cluster service name)
- **No Port Forwarding**: Direct cluster service communication

The example uses mirrord to run the HTTP server locally while intercepting traffic meant for the Kubernetes pod, allowing for rapid development and testing without rebuilding images. All communication between services uses Kubernetes service discovery.
