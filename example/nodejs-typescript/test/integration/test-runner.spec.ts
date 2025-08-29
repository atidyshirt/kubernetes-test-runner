import { expect } from 'chai';
import { TestContainer } from '../helper/testContainer';
import { getGlobalTestContainer, testEnvironment } from '../setup';
import 'mocha';

describe('Test Runner Integration Tests', () => {
  let testContainer: TestContainer;
  const namespace = testEnvironment.namespace;

  before(async () => {
    console.log(`Starting integration tests in namespace: ${namespace}`);
    testContainer = getGlobalTestContainer();
  });

  beforeEach(async () => {
    await testContainer.getMongoService().clearData();
  });

  it('test runner should have access to ket environment variables', () => {
    expect(process.env.KET_TEST_NAMESPACE).to.exist;
    expect(process.env.KET_PROJECT_ROOT).to.exist;
    expect(process.env.KET_WORKSPACE_PATH).to.exist;
    
    console.log('Test environment check:');
    console.log(`  Namespace: ${process.env.KET_TEST_NAMESPACE}`);
    console.log(`  Project Root: ${process.env.KET_PROJECT_ROOT}`);
    console.log(`  Workspace Path: ${process.env.KET_WORKSPACE_PATH}`);
  });

  it('should use the global test environment', () => {
    expect(testEnvironment.namespace).to.equal(namespace);
    
    console.log('Global test environment:');
    console.log(`  Namespace: ${testEnvironment.namespace}`);
    console.log(`  Project Root: ${testEnvironment.projectRoot}`);
    console.log(`  Workspace Path: ${testEnvironment.workspacePath}`);
  });
});
