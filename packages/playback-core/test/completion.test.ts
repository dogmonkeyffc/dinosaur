import assert from 'node:assert/strict';
import { test } from 'node:test';
import { completionPosition } from '../src/index';

test('native end duration handles stale final time-update events', () => {
  assert.equal(completionPosition(90, 100, 0, 100), 100);
  assert.equal(completionPosition(29, 30, 70, 100), 100);
  assert.equal(completionPosition(98, 0, 0, 100), 100);
});
test('truncated streams do not advance to the next episode', () => {
  assert.equal(completionPosition(20, 20, 0, 100), null);
  assert.equal(completionPosition(20, 20, 50, 100), null);
});
test('unknown media duration completes at its source end', () => {
  assert.equal(completionPosition(29, 30, 10, 0), 40);
});
