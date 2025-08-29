import { expect } from 'chai';
import axios from 'axios';
import { TestContainer } from '../helper/testContainer';
import { getGlobalTestContainer, testEnvironment } from '../setup';
import 'mocha';

describe('Example HTTP Server Integration Tests', () => {
  let testContainer: TestContainer;
  const namespace = testEnvironment.namespace;
  const baseUrl = `http://example-http-server.${namespace}.svc.cluster.local:3000`;

  before(async () => {
    console.log(`Starting integration tests in namespace: ${namespace}`);
    testContainer = getGlobalTestContainer();
  });

  beforeEach(async () => {
    await testContainer.getMongoService().clearData();
  });

  it('should write data to HTTP server and persist to MongoDB', async () => {
    const mongoService = testContainer.getMongoService();
    
    const testName = 'test-item';
    const testValue = 'test-value';

    const response = await axios.post(`${baseUrl}/data`, {
      name: testName,
      value: testValue
    });

    expect(response.status).to.equal(200);
    expect(response.data.success).to.be.true;
    expect(response.data.message).to.equal('Data inserted successfully');

    const savedData = await mongoService.getData(testName);
    expect(savedData).to.not.be.null;
    expect(savedData.name).to.equal(testName);
    expect(savedData.value).to.equal(testValue);
    expect(savedData.timestamp).to.exist;
  });

  it('should retrieve data from HTTP server', async () => {
    const mongoService = testContainer.getMongoService();
    
    await mongoService.insertData('retrieve-test', 'retrieve-value');

    const response = await axios.get(`${baseUrl}/data/retrieve-test`);
    
    expect(response.status).to.equal(200);
    expect(response.data.name).to.equal('retrieve-test');
    expect(response.data.value).to.equal('retrieve-value');
  });

  it('should return 404 for non-existent data', async () => {
    try {
      await axios.get(`${baseUrl}/data/non-existent`);
      expect.fail('Should have thrown an error');
    } catch (error: any) {
      expect(error.response.status).to.equal(404);
      expect(error.response.data.error).to.equal('Data not found');
    }
  });

  it('should return 400 for invalid data', async () => {
    try {
      await axios.post(`${baseUrl}/data`, {
        name: 'test'
      });
      expect.fail('Should have thrown an error');
    } catch (error: any) {
      expect(error.response.status).to.equal(400);
      expect(error.response.data.error).to.equal('Name and value are required');
    }
  });

  it('should handle health check endpoint', async () => {
    const response = await axios.get(`${baseUrl}/health`);
    
    expect(response.status).to.equal(200);
    expect(response.data.status).to.equal('healthy');
    expect(response.data.timestamp).to.exist;
  });

  it('should maintain data consistency across multiple operations', async () => {
    const mongoService = testContainer.getMongoService();
    
    const items = [
      { name: 'item1', value: 'value1' },
      { name: 'item2', value: 'value2' },
      { name: 'item3', value: 'value3' }
    ];

    for (const item of items) {
      await axios.post(`${baseUrl}/data`, item);
    }

    const count = await mongoService.countDocuments();
    expect(count).to.equal(3);

    for (const item of items) {
      const savedData = await mongoService.getData(item.name);
      expect(savedData.name).to.equal(item.name);
      expect(savedData.value).to.equal(item.value);
    }
  });
});
