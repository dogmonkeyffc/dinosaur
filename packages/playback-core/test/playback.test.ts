import assert from 'node:assert/strict';
import test from 'node:test';
import { ProgressQueue, sourcePosition } from '../src/index';
import { activeCue, parseVTT } from '../src/subtitles';

test('progress writes remain ordered after a failed request', async () => {
  const written: number[] = [];
  const queue = new ProgressQueue(async (position) => {
    written.push(position);
    if (position === 10) throw new Error('offline');
  });
  const failed = queue.save(10);
  const next = queue.save(20);
  await assert.rejects(failed);
  await next;
  assert.deepEqual(written, [10, 20]);
  await queue.save(NaN);
  assert.deepEqual(written, [10, 20]);
});

test('compatibility streams and subtitles use full source time', () => {
  assert.equal(sourcePosition(7, 120, 300), 127);
  assert.equal(sourcePosition(200, 120, 300), 300);
  const cues = parseVTT(
    'WEBVTT\n\n1\n00:02:00.000 --> 00:02:04.000\n<b>Hello</b> &amp; goodbye\n\n2\n00:02:05.000 --> 00:02:08.000 align:start\nNext',
  );
  assert.equal(activeCue(cues, 122), 'Hello & goodbye');
  assert.equal(activeCue(cues, 124), '');
  assert.equal(activeCue(cues, 125), 'Next');
});
