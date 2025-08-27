import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

export class KubectlServiceManager {
  private namespace: string;

  constructor(namespace: string) {
    this.namespace = namespace;
  }

  async applyManifest(manifestPath: string): Promise<void> {
    try {
      await execAsync(`kubectl apply -f ${manifestPath} -n ${this.namespace}`);
    } catch (error) {
      throw new Error(`Failed to apply manifest ${manifestPath}: ${error}`);
    }
  }

  async deleteManifest(manifestPath: string): Promise<void> {
    try {
      await execAsync(`kubectl delete -f ${manifestPath} -n ${this.namespace} --ignore-not-found=true`);
    } catch (error) {
      console.warn(`Warning: Failed to delete manifest ${manifestPath}: ${error}`);
    }
  }

  async waitForPodReady(podLabel: string, timeout: number = 60000): Promise<void> {
    const startTime = Date.now();
    
    while (Date.now() - startTime < timeout) {
      try {
        const { stdout } = await execAsync(
          `kubectl get pods -n ${this.namespace} -l ${podLabel} -o jsonpath='{.items[0].status.phase}'`
        );
        
        if (stdout.trim() === 'Running') {
          return;
        }
      } catch (error) {
        // Pod might not exist yet
      }
      
      await new Promise(resolve => setTimeout(resolve, 2000));
    }
    
    throw new Error(`Pod with label ${podLabel} not ready within ${timeout}ms`);
  }

  async waitForServiceReady(serviceName: string, timeout: number = 60000): Promise<void> {
    const startTime = Date.now();
    
    while (Date.now() - startTime < timeout) {
      try {
        const { stdout } = await execAsync(
          `kubectl get endpoints -n ${this.namespace} ${serviceName} -o jsonpath='{.subsets[0].addresses}'`
        );
        
        if (stdout.trim() !== '') {
          return;
        }
      } catch (error) {
        // Service might not exist yet
      }
      
      await new Promise(resolve => setTimeout(resolve, 2000));
    }
    
    throw new Error(`Service ${serviceName} not ready within ${timeout}ms`);
  }

  async getPodName(podLabel: string): Promise<string> {
    try {
      const { stdout } = await execAsync(
        `kubectl get pods -n ${this.namespace} -l ${podLabel} -o jsonpath='{.items[0].metadata.name}'`
      );
      return stdout.trim();
    } catch (error) {
      throw new Error(`Failed to get pod name for label ${podLabel}: ${error}`);
    }
  }

  async portForward(podName: string, localPort: number, containerPort: number): Promise<void> {
    try {
      await execAsync(
        `kubectl port-forward -n ${this.namespace} ${podName} ${localPort}:${containerPort}`
      );
    } catch (error) {
      throw new Error(`Failed to setup port forward for ${podName}: ${error}`);
    }
  }
}
