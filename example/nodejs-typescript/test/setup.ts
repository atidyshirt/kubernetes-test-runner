import { TestContainer } from './helper/testContainer';
import 'mocha';

const testEnvironment = {
  namespace: process.env.KET_TEST_NAMESPACE!,
  projectRoot: process.env.KET_PROJECT_ROOT!,
  workspacePath: process.env.KET_WORKSPACE_PATH!
};

let globalTestContainer: TestContainer;

export function getGlobalTestContainer(): TestContainer {
  if (!globalTestContainer) {
    globalTestContainer = new TestContainer(testEnvironment.namespace);
  }
  return globalTestContainer;
}

export { testEnvironment };

// Initialize the test container when this file is loaded
globalTestContainer = new TestContainer(testEnvironment.namespace);

// Perform any global setup
console.log('Setting up test environment');
console.log(`Namespace: ${testEnvironment.namespace}`);
console.log(`Project Root: ${testEnvironment.projectRoot}`);
console.log(`Workspace Path: ${testEnvironment.workspacePath}`);

// Mocha global hooks
before(async () => {
  await globalTestContainer.setup();
});

after(async () => {
  await globalTestContainer.teardown();
});
