import { before, after } from 'mocha';

before(async function () {
  console.log('Global before: do no work');
});

after(async function () {
  console.log('Global after: do no work');
});
