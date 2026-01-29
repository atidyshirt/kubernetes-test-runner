import { KubectlServiceManager } from './kubectl-service-manager.ts';
import { WiremockService } from './wiremock-service.ts';

export class TestContainer {
    readonly kubectl: KubectlServiceManager;
    private wiremockService: WiremockService;

    constructor() {
        this.kubectl = new KubectlServiceManager();
        this.wiremockService = new WiremockService();
    }

    async countStringInWorkloadLogs(workload: string, search: string): Promise<number> {
        return this.kubectl.countStringInWorkloadLogs(workload, search);
    }

    async waitForWorkloadLogCount(
        workload: string,
        search: string,
        target: number,
    ): Promise<void> {
        return this.kubectl.waitForWorkloadLogCount(workload, search, target);
    }

    async updateMapping(
        serviceName: string,
        endpoint: string,
        method: string,
        jsonBody: object,
    ): Promise<void> {
        await this.wiremockService.updateMapping(serviceName, endpoint, method, jsonBody);
    }

    async resetMappings(serviceName: string): Promise<void> {
        await this.wiremockService.resetMappings(serviceName);
    }

    async getRequests(serviceName: string): Promise<any[]> {
        return this.wiremockService.getRequests(serviceName);
    }

    async resetRequests(serviceName: string): Promise<void> {
        await this.wiremockService.resetRequests(serviceName);
    }
}
