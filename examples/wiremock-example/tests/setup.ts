import { before, after } from 'mocha';

before(async function () {
  console.log('Global before: setup started');
});

after(async function () {
  console.log('Global after: teardown started');
});
