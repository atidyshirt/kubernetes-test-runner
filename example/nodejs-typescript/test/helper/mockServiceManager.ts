import { spawn, ChildProcess } from 'child_process';
import { KubectlServiceManager } from './kubectlServiceManager';

export class MockServiceManager {
  private kubectl: KubectlServiceManager;
  private mirrordProcess: ChildProcess | null = null;

  constructor(namespace: string) {
    this.kubectl = new KubectlServiceManager(namespace);
  }

  async startMockService(podLabel: string): Promise<void> {
    try {
      const podName = await this.kubectl.getPodName(podLabel);
      
      this.mirrordProcess = spawn('mirrord', [
        'exec',
        '--steal',
        '--target', podName,
        '--target-namespace', this.kubectl['namespace'],
        '--', 'node', 'src/example-http-server.ts'
      ], {
        stdio: 'pipe',
        cwd: process.cwd()
      });

      await new Promise(resolve => setTimeout(resolve, 5000));
    } catch (error) {
      throw new Error(`Failed to start mock service: ${error}`);
    }
  }

  async stopMockService(): Promise<void> {
    try {
      if (this.mirrordProcess) {
        this.mirrordProcess.kill('SIGTERM');
        this.mirrordProcess = null;
      }
    } catch (error) {
      console.warn('Warning: Failed to stop mock service:', error);
    }
  }

  getMirrordProcess(): ChildProcess | null {
    return this.mirrordProcess;
  }

  isRunning(): boolean {
    return this.mirrordProcess !== null && !this.mirrordProcess.killed;
  }
}
