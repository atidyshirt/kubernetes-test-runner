# Node.js + TypeScript Integration Test Example

This example demonstrates how to use ket for running integration tests with Node.js, TypeScript, and MongoDB.

## Structure

- `src/example-http-server.ts` - Express server that writes to MongoDB
- `manifests/` - Kubernetes manifests for MongoDB and HTTP server
- `test/helper/` - Test utilities for managing cluster state and services
- `test/integration/` - Integration tests that verify the full flow

## Quick Start

1. Install dependencies:
   ```bash
   npm install
   ```

2. Build TypeScript:
   ```bash
   npm run build
   ```

3. Start Kind cluster:
   ```bash
   npm run cluster:start
   ```

4. Fix RBAC permissions:
   ```bash
   npm run cluster:fix-rbac
   ```

5. Run integration tests:
   ```bash
   npm run test:integration
   ```

## Available Scripts

### Testing
- `npm run test:integration` - Run integration tests via ket
- `npm run test:integration:watch` - Run tests with file watching
- `npm run test:unit` - Run unit tests directly

### Cluster Management
- `npm run cluster:start` - Create Kind cluster
- `npm run cluster:stop` - Delete Kind cluster
- `npm run cluster:restart` - Restart cluster
- `npm run cluster:status` - Show cluster status
- `npm run cluster:create-ns` - Create test namespace
- `npm run cluster:fix-rbac` - Fix RBAC permissions

### Development
- `npm run build` - Compile TypeScript
- `npm run deploy` - Build Docker image
- `npm run start` - Run compiled server
- `npm run dev` - Run server with ts-node

## Workflow

1. **Setup**: Install dependencies and build TypeScript
2. **Cluster**: Start Kind cluster and configure RBAC
3. **Test**: Run integration tests via ket
4. **Cleanup**: Stop cluster when done

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

The example uses mirrord to run the HTTP server locally while intercepting traffic meant for the Kubernetes pod, allowing for rapid development and testing without rebuilding images.
