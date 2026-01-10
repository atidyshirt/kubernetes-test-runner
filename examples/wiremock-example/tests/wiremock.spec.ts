import axios from 'axios';
import { expect } from 'chai';
import { TestContainer } from 'kubernetes-embedded-testing';

describe('wiremock example', function () {
  let testContainer: TestContainer;

  before(async function () {
    testContainer = new TestContainer('test');
    await testContainer.setup();
  });

  after(async function () {
    await testContainer.teardown();
  });

  it('should respond to admin endpoint (cluster-only)', async function () {
    const res = await axios.get('http://wiremock:8080/__admin/');
    expect(res.status).equal(200);
  });

  it('should return mocked data through example service', async function () {
    await testContainer.updateMapping('wiremock', '/api/data', 'GET', { message: 'Hello from Wiremock!' });

    const res = await axios.get('http://example-service:3000/api/data');
    expect(res.status).equal(200);
    expect(res.data).to.deep.equal({ message: 'Hello from Wiremock!' });
  });
});
