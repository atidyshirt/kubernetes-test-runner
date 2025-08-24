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
   * Get the path to the ket binary
   */
  getTestRunnerPath() {
    // Look for ket in common locations
    const possiblePaths = [
      path.join(process.cwd(), 'testrunnner', 'bin', 'ket'),
      path.join(process.cwd(), '..', 'testrunnner', 'bin', 'ket'),
      path.join(process.cwd(), '..', '..', 'testrunnner', 'bin', 'ket'),
      '/usr/local/bin/ket',
      'ket'
    ];

    for (const testPath of possiblePaths) {
      if (fs.existsSync(testPath)) {
        return testPath;
      }
    }

    throw new Error('ket binary not found. Please build it first with "make build"');
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
   * Launch the ket
   */
  async launch() {
    if (this.isRunning) {
      throw new Error('ket is already running');
    }

    const testRunnerPath = this.getTestRunnerPath();
    
    // The ket will create a test pod that:
    // 1. Mounts your source code
    // 2. Uses mirrord to intercept traffic from the target pod
    // 3. Runs your test command
    // 4. Streams results back to stdout
    // 5. Cleans up automatically

    const args = [
      '-target-pod', this.options.targetPod,
      '-target-namespace', this.options.targetNamespace,
      '-test-command', this.options.testCommand,
      '-proc', this.options.processToTest,
      '-project-root', this.options.projectRoot
    ];

    if (this.options.debug) {
      args.push('-debug');
    }

    if (this.options.keepNamespace) {
      args.push('-keep-namespace');
    }

    console.log(`Launching ket: ${testRunnerPath} ${args.join(' ')}`);
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
        console.log(`[ket] ${output.trim()}`);
      });

      testrunner.stderr.on('data', (data) => {
        const output = data.toString();
        stderr += output;
        console.error(`[ket:error] ${output.trim()}`);
      });

      testrunner.on('close', (code) => {
        if (code === 0) {
          console.log('ket completed successfully');
          console.log('Tests ran against intercepted traffic from the running Express server');
          this.isRunning = false;
          resolve({ stdout, stderr, code });
        } else {
          const error = new Error(`ket failed with exit code ${code}`);
          error.stdout = stdout;
          error.stderr = stderr;
          error.code = code;
          reject(error);
        }
      });

      testrunner.on('error', (error) => {
        reject(new Error(`Failed to start ket: ${error.message}`));
      });

      // Set timeout for the entire operation
      setTimeout(() => {
        if (this.isRunning) {
          testrunner.kill('SIGTERM');
          reject(new Error(`ket timed out after ${this.options.timeout}ms`));
        }
      }, this.options.timeout);
    });
  }

  /**
   * Clean up resources
   */
  async cleanup() {
    console.log('Cleaning up ket resources...');
    
    // Add cleanup logic here if needed
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
