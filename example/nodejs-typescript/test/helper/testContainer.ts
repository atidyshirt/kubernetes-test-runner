import { KubectlServiceManager } from './kubectlServiceManager';
import { MongoService } from './mongoService';
import { MockServiceManager } from './mockServiceManager';

export class TestContainer {
  private kubectl: KubectlServiceManager;
  private mongo: MongoService;
  private mockService: MockServiceManager;
  private namespace: string;

  constructor(namespace: string) {
    this.namespace = namespace;
    this.kubectl = new KubectlServiceManager(namespace);
    this.mongo = new MongoService(`mongodb://mongodb.${namespace}.svc.cluster.local:27017`);
    this.mockService = new MockServiceManager(namespace);
  }

  async setup(): Promise<void> {
    try {
      await this.kubectl.applyManifest('manifests/mongodb.yml');
      await this.kubectl.waitForPodReady('app=mongodb');
      
      await this.kubectl.applyManifest('manifests/example-http-server.yml');
      await this.kubectl.waitForPodReady('app=example-http-server');
      
      await this.kubectl.waitForServiceReady('mongodb');
      await this.mongo.connect();
      
      await this.mockService.startMockService('app=example-http-server');
    } catch (error) {
      throw error;
    }
  }

  async teardown(): Promise<void> {
    try {
      await this.mockService.stopMockService();
      await this.mongo.disconnect();
      
      await this.kubectl.deleteManifest('manifests/example-http-server.yml');
      await this.kubectl.deleteManifest('manifests/mongodb.yml');
    } catch (error) {
      console.warn('Warning: Failed to teardown test environment:', error);
    }
  }

  getMongoService(): MongoService {
    return this.mongo;
  }

  getMockService(): MockServiceManager {
    return this.mockService;
  }

  getKubectlService(): KubectlServiceManager {
    return this.kubectl;
  }
}
