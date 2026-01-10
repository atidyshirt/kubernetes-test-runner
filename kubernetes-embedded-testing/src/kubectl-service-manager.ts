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
            await execAsync(
                `kubectl delete -f ${manifestPath} -n ${this.namespace} --ignore-not-found=true`,
            );
        } catch (error) {
            console.warn(`Warning: Failed to delete manifest ${manifestPath}: ${error}`);
        }
    }

    async waitForPodReady(podLabel: string): Promise<void> {
        const command = `kubectl wait --for=condition=Ready pod -l ${podLabel} -n ${this.namespace} --timeout=500s`;
        console.log(command);
        await execAsync(command);
    }

    async killPod(podLabel: string): Promise<void> {
        const command = `kubectl delete pod -l ${podLabel} -n ${this.namespace} --force --grace-period=0`;
        console.log(command);
        await execAsync(command);
    }

    async waitForWorkloadToBeReady(deployment: string): Promise<void> {
        const command = `kubectl rollout status ${deployment} -n ${this.namespace} -w --timeout=90s`;
        await execAsync(command);
        console.log(command);
    }

    async countStringInWorkloadLogs(workload: string, search: string): Promise<number> {
        const cmd = `kubectl logs -n ${this.namespace} ${workload} | grep -o "${search}" | wc -l`;
        const { stdout } = await execAsync(cmd);
        return Number(stdout.trim()) || 0;
    }

    async waitForWorkloadLogCount(
        workload: string,
        search: string,
        target: number,
    ): Promise<void> {
        while (true) {
            const count = await this.countStringInWorkloadLogs(workload, search);
            if (count >= target) {
                break;
            }
            await new Promise((r) => setTimeout(r, 10));
        }
    }

    async getPodName(podLabel: string): Promise<string> {
        try {
            const { stdout } = await execAsync(
                `kubectl get pods -n ${this.namespace} -l ${podLabel} -o jsonpath='{.items[0].metadata.name}'`,
            );
            return stdout.trim();
        } catch (error) {
            throw new Error(`Failed to get pod name for label ${podLabel}: ${error}`);
        }
    }
}
