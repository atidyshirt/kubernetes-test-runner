# Kubernetes Test Runner

A powerful integration testing framework for Kubernetes that allows you to run tests against pods while using mirrord for traffic interception. This tool creates isolated test environments, mounts your local source code, and executes tests with full access to your Kubernetes cluster.

## Features

- 🚀 **Easy Launch**: Simple CLI to launch test jobs in Kubernetes
- 📁 **Source Code Mounting**: Automatically mounts local directory contents via ConfigMaps
- 🔄 **Mirrord Integration**: Uses mirrord to intercept traffic to your target pod
- 🧪 **Flexible Testing**: Support for any test framework (Mocha, Jest, Go tests, etc.)
- 📊 **Comprehensive Reporting**: JSON test results with timing and metadata
- 🧹 **Auto Cleanup**: Automatic namespace cleanup after test completion
- 🔧 **Customizable**: Configurable timeouts, retry limits, and resource settings

## Architecture

The testrunner operates in two modes:

1. **Launch Mode** (`testrunner launch`): Creates a Kubernetes job that runs your tests
2. **Run Mode** (`testrunner run`): Executes inside the Kubernetes container to run tests

```
Local Development → testrunner launch → Kubernetes Job → testrunner run → Test Execution
```

## Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd kubernetes-test-runner

# Build the binary
make build

# Install to system (optional)
make install
```

### Option 2: Use Docker

```bash
# Build Docker image
make docker-build

# Run with Docker
docker run --rm -v $(pwd):/workspace testrunner:latest --help
```

## Quick Start

### 1. Basic Usage

```bash
# Launch a test job
testrunner launch \
  --target-pod my-app-pod \
  --target-namespace default \
  --test-command "npm test" \
  --proc "npm start"
```

### 2. With Custom Test Command

```bash
# Run Mocha tests
testrunner launch \
  --target-pod my-app-pod \
  --target-namespace default \
  --test-command "mocha **/*.spec.js" \
  --proc "npm run start"
```

### 3. Go Tests

```bash
# Run Go tests
testrunner launch \
  --target-pod my-go-app \
  --target-namespace default \
  --test-command "go test ./..." \
  --proc "./my-app"
```

## Examples

### Node.js Application Testing

```bash
# Test a Node.js app with Mocha
testrunner launch \
  --target-pod my-nodejs-app \
  --target-namespace production \
  --test-command "npm install && npm test" \
  --proc "npm start" \
  --image node:18-alpine \
  --debug
```

### Go Application Testing

```bash
# Test a Go app
testrunner launch \
  --target-pod my-go-app \
  --target-namespace staging \
  --test-command "go mod download && go test -v ./..." \
  --proc "./my-app" \
  --image golang:1.21-alpine
```

### Python Application Testing

```bash
# Test a Python app with pytest
testrunner launch \
  --target-pod my-python-app \
  --target-namespace development \
  --test-command "pip install -r requirements.txt && pytest" \
  --proc "python app.py" \
  --image python:3.11-alpine
```

## Advanced Examples

### Mocha + Node.js Integration

For a complete example showing how to integrate TestRunner with Mocha testing, see the `example/mocha-nodejs` directory. This example demonstrates:

- Express server running in Kubernetes
- Mocha hooks managing TestRunner lifecycle
- Unit and integration tests
- Clean `npm test` workflow

```bash
# Setup the example
cd example/mocha-nodejs
./setup.sh

# Run all tests
npm test
```

## Command Line Options

### Global Options

| Option | Description | Default |
|--------|-------------|---------|
| `--mode` | Operation mode: `launch` or `run` | `launch` |
| `--debug` | Enable debug logging | `false` |

### Launch Mode Options

| Option | Description | Default | Required |
|--------|-------------|---------|----------|
| `--target-pod` | Target pod to test against | - | ✅ |
| `--target-namespace` | Target namespace | `default` | ✅ |
| `--test-command` | Test command to execute | - | ✅ |
| `--proc` | Process to test against | - | ✅ |
| `--project-root` | Project root path | `.` | ❌ |
| `--image` | Runner image | `node:18-alpine` | ❌ |
| `--namespace` | Test namespace | `testrunner-{project}` | ❌ |
| `--keep-namespace` | Keep test namespace after run | `false` | ❌ |
| `--test-timeout` | Test timeout in seconds | `300` | ❌ |
| `--active-deadline-seconds` | Job deadline in seconds | `1800` | ❌ |
| `--backoff-limit` | Job backoff limit | `1` | ❌ |

## How It Works

### 1. Source Code Mounting

The testrunner creates a ConfigMap from your local directory and mounts it into the test container:

```yaml
volumes:
- name: source-code
  configMap:
    name: testrunner-source-{project}
```

### 2. Mirrord Integration

The test container downloads and runs mirrord to intercept traffic:

```bash
/tools/mirrord exec --target-pod {pod} --target-namespace {ns} -- {process} &
```

### 3. Test Execution

Your test command runs in the same container with access to:
- Your source code in `/workspace`
- Mirrord for traffic interception
- Full Kubernetes cluster access

### 4. Result Collection

Test results are written to `/reports/test-results.json` with:
- Success/failure status
- Execution time
- Exit codes
- Test metadata

## Configuration

### Environment Variables

The test container receives these environment variables:

- `TARGET_NAMESPACE`: Target namespace for mirrord
- `TARGET_POD`: Target pod for mirrord
- `PROCESS_TO_TEST`: Process command to test
- `TEST_COMMAND`: Test command to execute

### Custom Images

You can specify custom runner images for different language environments:

```bash
# Node.js
--image node:18-alpine

# Go
--image golang:1.21-alpine

# Python
--image python:3.11-alpine

# Custom
--image my-registry.com/testrunner:latest
```

## Advanced Usage

### Persistent Namespaces

Keep test namespaces for debugging:

```bash
testrunner launch \
  --target-pod my-app \
  --test-command "npm test" \
  --proc "npm start" \
  --keep-namespace
```

### Custom Timeouts

Set custom timeouts for long-running tests:

```bash
testrunner launch \
  --target-pod my-app \
  --test-command "npm run test:integration" \
  --proc "npm start" \
  --test-timeout 600 \
  --active-deadline-seconds 3600
```

### Debug Mode

Enable verbose logging:

```bash
testrunner launch \
  --target-pod my-app \
  --test-command "npm test" \
  --proc "npm start" \
  --debug
```

## Troubleshooting

### Common Issues

1. **Pod Not Found**: Ensure the target pod exists and is running
2. **Permission Denied**: Check RBAC permissions for the test namespace
3. **Test Timeout**: Increase `--test-timeout` and `--active-deadline-seconds`
4. **Mirrord Download Failed**: Check network connectivity and mirrord URL

### Debug Information

Enable debug mode to see:
- Configuration details
- File mounting information
- Container execution steps
- Detailed error messages

### Logs

View test execution logs:

```bash
# Get pod logs
kubectl logs -f job/testrunner-{project} -n {namespace}

# Get job status
kubectl get job testrunner-{project} -n {namespace}
```

## Development

### Building

```bash
# Build binary
make build

# Build for multiple platforms
make build-all

# Build Docker image
make docker-build
```

### Testing

```bash
# Run unit tests
make test

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Lint code
make lint
```

### Project Structure

```
kubernetes-test-runner/
├── testrunnner/                    # Main Go application
│   ├── cmd/testrunner/            # CLI entry point
│   ├── pkg/
│   │   ├── config/                # Configuration and CLI flags
│   │   ├── kube/                  # Kubernetes operations
│   │   ├── launcher/              # Job launch logic
│   │   ├── runner/                # Test execution logic
│   │   └── report/                # Test result reporting
│   ├── Dockerfile                 # Container image definition
│   └── Makefile                   # Build and development tasks
├── example/                        # Example applications and tests
│   └── mocha-nodejs/              # Mocha + Node.js integration example
├── demo.sh                        # Interactive demonstration script
└── README.md                      # This file
```

## Contributing

For issues and questions:
- Create an issue in the repository
- Check the troubleshooting section
- Review the debug logs with `--debug` flag

## License

[Add your license information here]

## Support

For issues and questions:
- Create an issue in the repository
- Check the troubleshooting section
- Review the debug logs with `--debug` flag
