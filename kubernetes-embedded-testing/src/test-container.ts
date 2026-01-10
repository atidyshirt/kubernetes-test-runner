import { KubectlServiceManager } from './kubectl-service-manager.ts';
import { MongoService } from './mongo-service.ts';
import { WiremockService } from './wiremock-service.ts';

export class TestContainer {
    readonly kubectl: KubectlServiceManager;
    private mongoService: MongoService;
    private wiremockService: WiremockService;

    constructor(namespace: string) {
        this.kubectl = new KubectlServiceManager(namespace);
        this.mongoService = new MongoService(
            `mongodb://mongodb.${namespace}.svc.cluster.local:27017`,
        );
        this.wiremockService = new WiremockService(namespace);
    }

    private async resetMappings() {
        await this.wiremockService.resetMappings('cpu-service');
        await this.wiremockService.resetMappings('author-service');
    }

    async setup(): Promise<void> {
        try {
            await this.resetMappings();
            await this.mongoService.connect();
        } catch (error) {
            throw error;
        }
    }

    async teardown(): Promise<void> {
        try {
            await this.resetMappings();
        } catch (error) {
            console.warn(`Warning: Failed to teardown test environment: ${error}`);
        }
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

    getDb() {
        return this.mongoService.getDb();
    }

    async updateMapping(
        serviceName: string,
        endpoint: string,
        method: string,
        jsonBody: object,
    ): Promise<void> {
        await this.wiremockService.updateMapping(serviceName, endpoint, method, jsonBody);
    }

    async afterEach(): Promise<void> {
        await this.resetMappings();
    }
}
