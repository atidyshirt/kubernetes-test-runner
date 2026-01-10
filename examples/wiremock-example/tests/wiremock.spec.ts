import axios from 'axios';
import { expect } from 'chai';

describe('wiremock example', function () {
  it('should respond to admin endpoint (cluster-only)', async function () {
    const res = await axios.get('http://wiremock:8080/__admin/');
    expect(res.status).equal(200);
  });
});
