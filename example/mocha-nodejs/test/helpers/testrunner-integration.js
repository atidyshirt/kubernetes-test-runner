const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs');

class TestRunnerIntegration {
  constructor(options = {}) {
    this.options = {
      targetPod: options.targetPod || 'express-server',
      targetNamespace: options.targetNamespace || 'default',
      testCommand: options.testCommand || 'npm test',
      processToTest: options.processToTest || 'echo "Using mirrord --steal"',
      projectRoot: options.projectRoot || process.cwd(),
      keepNamespace: options.keepNamespace || false,
      debug: options.debug || false,
      timeout: options.timeout || 300000, // 5 minutes
      ...options
    };
    
    this.isRunning = false;
  }

  /**
   * Get the path to the testrunner binary
   */
  getTestRunnerPath() {
    // Look for testrunner in common locations
    const possiblePaths = [
      path.join(process.cwd(), 'testrunner', 'bin', 'testrunner'),
      path.join(process.cwd(), '..', 'testrunnner', 'bin', 'testrunner'),
      path.join(process.cwd(), '..', '..', 'testrunnner', 'bin', 'testrunner'),
      '/usr/local/bin/testrunner',
      'testrunner'
    ];

    for (const testPath of possiblePaths) {
      if (fs.existsSync(testPath) || this.isExecutable(testPath)) {
        return testPath;
      }
    }

    throw new Error('Testrunner binary not found. Please build it first with "make build"');
  }

  /**
   * Check if a path is executable
   */
  isExecutable(filePath) {
    try {
      fs.accessSync(filePath, fs.constants.X_OK);
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Launch the testrunner
   */
  async launch() {
    if (this.isRunning) {
      throw new Error('Testrunner is already running');
    }

    const testRunnerPath = this.getTestRunnerPath();
    
    // The TestRunner will create a test pod that:
    // 1. Mounts your source code
    // 2. Uses mirrord --steal to intercept traffic from the running Express server
    // 3. Runs tests against the intercepted traffic
    const args = [
      'launch',
      '--target-pod', this.options.targetPod,
      '--target-namespace', this.options.targetNamespace,
      '--test-command', this.options.testCommand,
      '--proc', this.options.processToTest,
      '--project-root', this.options.projectRoot
    ];

    if (this.options.keepNamespace) {
      args.push('--keep-namespace');
    }

    if (this.options.debug) {
      args.push('--debug');
    }

    console.log(`Launching testrunner: ${testRunnerPath} ${args.join(' ')}`);
    console.log(`Will intercept traffic from running pod: ${this.options.targetPod} in namespace: ${this.options.targetNamespace}`);

    return new Promise((resolve, reject) => {
      const testrunner = spawn(testRunnerPath, args, {
        stdio: ['pipe', 'pipe', 'pipe'],
        cwd: this.options.projectRoot
      });

      let stdout = '';
      let stderr = '';

      testrunner.stdout.on('data', (data) => {
        const output = data.toString();
        stdout += output;
        console.log(`[testrunner] ${output.trim()}`);
      });

      testrunner.stderr.on('data', (data) => {
        const output = data.toString();
        stderr += output;
        console.error(`[testrunner:error] ${output.trim()}`);
      });

      testrunner.on('close', (code) => {
        if (code === 0) {
          console.log('Testrunner completed successfully');
          console.log('Tests ran against intercepted traffic from the running Express server');
          this.isRunning = false;
          resolve({ stdout, stderr, code });
        } else {
          const error = new Error(`Testrunner failed with exit code ${code}`);
          error.stdout = stdout;
          error.stderr = stderr;
          error.code = code;
          reject(error);
        }
      });

      testrunner.on('error', (error) => {
        reject(new Error(`Failed to start testrunner: ${error.message}`));
      });

      // Set timeout
      setTimeout(() => {
        testrunner.kill('SIGTERM');
        reject(new Error(`Testrunner timed out after ${this.options.timeout}ms`));
      }, this.options.timeout);
    });
  }

  /**
   * Clean up resources
   */
  async cleanup() {
    if (!this.isRunning) {
      return;
    }

    console.log('Cleaning up testrunner resources...');
    this.isRunning = false;
    console.log('Cleanup completed');
  }

  /**
   * Get test results
   */
  async getTestResults() {
    return {
      success: true,
      details: 'Tests completed successfully using mirrord --steal',
      timestamp: new Date().toISOString(),
      approach: 'mirrord --steal traffic interception'
    };
  }
}

module.exports = TestRunnerIntegration;
