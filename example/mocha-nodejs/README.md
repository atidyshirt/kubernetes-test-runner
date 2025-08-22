# Mocha + Node.js TestRunner Integration Example

This example demonstrates how to integrate the Kubernetes TestRunner framework with a Node.js Express application using Mocha for testing. The example shows how to use Mocha hooks to manage the TestRunner lifecycle and test a server running in Kubernetes with mirrord traffic interception.

## 🏗️ Architecture

```
Local Development → Mocha Test → TestRunner Integration → Kubernetes Job → Test Execution
```

1. **Mocha Test Suite**: Defines integration tests using Mocha hooks
2. **TestRunner Integration**: Node.js module that interfaces with the testrunner binary
3. **Kubernetes Deployment**: Express server runs as a pod in the cluster
4. **Mirrord Integration**: Traffic interception for testing the server
5. **Test Execution**: Tests run in isolated Kubernetes environment

## 📁 Project Structure

```
mocha-nodejs/
├── src/
│   └── server.js                 # Express server implementation
├── test/
│   ├── helpers/
│   │   └── testrunner-integration.js  # TestRunner integration helper
│   ├── integration/
│   │   └── express-server.spec.js     # Integration tests with TestRunner
│   └── unit/
│       └── server.spec.js             # Unit tests (local execution)
├── manifests/
│   └── express-server-pod.yaml        # Kubernetes pod manifest
├── Dockerfile                          # Container image for Express server
├── package.json                        # Node.js dependencies and scripts
└── README.md                           # This file
```

## 🚀 Quick Start

### Prerequisites

- Node.js 18+ and npm
- Kubernetes cluster (local or remote)
- kubectl configured
- TestRunner binary built (see main README)

### 1. Install Dependencies

```bash
cd example/mocha-nodejs
npm install
```

### 2. Build TestRunner Binary

```bash
# From the project root
cd testrunnner
make build
cd ../example/mocha-nodejs
```

### 3. Deploy Express Server

```bash
# Apply the Kubernetes manifest
kubectl apply -f manifests/express-server-pod.yaml

# Wait for the pod to be ready
kubectl wait --for=condition=ready pod/express-server --timeout=60s
```

### 4. Run Tests

```bash
# Run all tests (unit + integration)
npm test

# Run tests in watch mode
npm run test:watch
```

That's it! The `npm test` command will:
1. Run unit tests locally
2. Launch TestRunner to run integration tests in Kubernetes
3. Report results back to Mocha

## 🧪 Test Types

### Unit Tests (`test/unit/server.spec.js`)

- **Purpose**: Test Express server functionality locally
- **Execution**: Runs in local Node.js environment
- **Coverage**: API endpoints, validation, error handling
- **Dependencies**: supertest for HTTP testing

### Integration Tests (`test/integration/express-server.spec.js`)

- **Purpose**: Test server in Kubernetes environment via TestRunner
- **Execution**: Uses TestRunner to launch Kubernetes job
- **Coverage**: End-to-end testing with mirrord traffic interception
- **Dependencies**: TestRunner integration helper

## 🔧 TestRunner Integration

The `TestRunnerIntegration` class provides a clean interface to the testrunner binary:

```javascript
const TestRunnerIntegration = require('./helpers/testrunner-integration');

const testRunner = new TestRunnerIntegration({
  targetPod: 'express-server',
  targetNamespace: 'default',
  testCommand: 'npm test',
  processToTest: 'npm start',
  projectRoot: process.cwd(),
  debug: true
});

// Launch TestRunner - this runs the tests and completes
await testRunner.launch();
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `targetPod` | Target pod to test against | `express-server` |
| `targetNamespace` | Target namespace | `default` |
| `testCommand` | Test command to execute | `npm test` |
| `processToTest` | Process to test against | `npm start` |
| `projectRoot` | Project root path | `process.cwd()` |
| `keepNamespace` | Keep test namespace | `false` |
| `debug` | Enable debug logging | `false` |
| `timeout` | Launch timeout in ms | `300000` |

## 🎯 Mocha Hooks Integration

The integration tests use Mocha's `before` and `after` hooks to manage the TestRunner lifecycle:

```javascript
describe('Express Server Integration Tests with TestRunner', function() {
  let testRunner;
  
  before(async function() {
    // Setup and launch TestRunner
    testRunner = new TestRunnerIntegration({...});
    await testRunner.launch();
  });
  
  after(async function() {
    // Cleanup
    await testRunner.cleanup();
  });
  
  // Test cases...
});
```

## 🌐 Express Server Features

The example Express server includes:

- **Health Check**: `/health` endpoint for Kubernetes probes
- **API Documentation**: Root endpoint with available routes
- **Users CRUD**: Full CRUD operations for user management
- **Validation**: Input validation and error handling
- **Security**: Helmet middleware for security headers
- **Logging**: Morgan for HTTP request logging

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check for Kubernetes |
| GET | `/` | API information and documentation |
| GET | `/api/users` | List all users |
| GET | `/api/users/:id` | Get user by ID |
| POST | `/api/users` | Create new user |
| PUT | `/api/users/:id` | Update existing user |
| DELETE | `/api/users/:id` | Delete user |

## 🐳 Containerization

### Building the Image

```bash
# Build Express server image
npm run docker:build

# Run locally
npm run docker:run
```

### Kubernetes Deployment

The `manifests/express-server-pod.yaml` includes:

- Pod with Express server container
- Service for network access
- Health checks and readiness probes
- Resource limits and requests
- Volume mounts for application code

## 🔍 Debugging

### Enable Debug Mode

```bash
# Set debug environment variable
export DEBUG=true

# Run tests with debug output
npm test
```

### TestRunner Debug Output

Debug mode provides detailed information about:

- TestRunner binary location
- Launch command and arguments
- Kubernetes job creation
- Test execution progress

### Kubernetes Debugging

```bash
# Check pod status
kubectl get pods -l app=express-server

# View pod logs
kubectl logs express-server

# Check pod events
kubectl describe pod express-server
```

## 🚨 Troubleshooting

### Common Issues

1. **TestRunner Binary Not Found**
   ```bash
   # Ensure binary is built
   cd testrunnner && make build
   ```

2. **Pod Not Ready**
   ```bash
   # Check pod status and events
   kubectl describe pod express-server
   kubectl logs express-server
   ```

3. **Test Timeout**
   ```bash
   # Increase timeout in test configuration
   this.timeout(300000); // 5 minutes
   ```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DEBUG` | Enable debug mode | `false` |
| `KEEP_NAMESPACE` | Keep test namespace | `false` |
| `NODE_ENV` | Node.js environment | `development` |

## 📊 Test Results

The TestRunner generates comprehensive test reports including:

- Test execution status (success/failure)
- Execution timing and duration
- Exit codes and error details
- Target pod and namespace information
- Test command and process details

## 🔄 CI/CD Integration

This example can be easily integrated into CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run Tests
  run: |
    cd example/mocha-nodejs
    npm install
    npm test
  env:
    DEBUG: true
```

## 🎉 Next Steps

1. **Customize Tests**: Modify test cases for your specific application
2. **Extend Server**: Add more endpoints and functionality
3. **Scale Testing**: Run multiple test suites concurrently
4. **Production Deployment**: Adapt manifests for production clusters
5. **Monitoring**: Add metrics and observability

## 📚 Additional Resources

- [TestRunner Framework Documentation](../README.md)
- [Mocha Testing Framework](https://mochajs.org/)
- [Express.js Documentation](https://expressjs.com/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Mirrord Documentation](https://mirrord.dev/)
