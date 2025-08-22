#!/bin/bash

# Simple setup script for Mocha + Node.js TestRunner example
# This script handles basic prerequisites and setup

set -e

echo "🔧 Setting up Mocha + Node.js TestRunner example..."

# Install dependencies
echo "📦 Installing Node.js dependencies..."
npm install

# Build testrunner binary if it doesn't exist
if [ ! -f "../../testrunnner/bin/testrunner" ]; then
    echo "🔨 Building testrunner binary..."
    cd ../../testrunnner
    make build
    cd ../example/mocha-nodejs
else
    echo "✅ Testrunner binary already exists"
fi

# Build Express server Docker image
echo "🐳 Building Express server Docker image..."
npm run docker:build

# Deploy Express server
echo "🚀 Deploying Express server to Kubernetes..."
kubectl apply -f manifests/express-server-pod.yaml

# Wait for deployment to be ready
echo "⏳ Waiting for Express server deployment to be ready..."
kubectl wait --for=condition=available deployment/express-server --timeout=120s

# Wait for pod to be ready
echo "⏳ Waiting for Express server pod to be ready..."
kubectl wait --for=condition=ready pod -l app=express-server --timeout=120s

echo "✅ Setup completed successfully!"
echo ""
echo "You can now run: npm test"
echo ""
echo "The TestRunner will:"
echo "1. Create a test pod with your source code mounted"
echo "2. Use mirrord --steal to intercept traffic from the running Express server"
echo "3. Run tests against the intercepted traffic"
echo ""
echo "The Express server runs independently as a normal Docker container!"
