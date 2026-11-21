import assert from 'node:assert/strict';
import test from 'node:test';
import {
  bufferedAhead,
  MetricSample,
  playbackMetricID,
  playbackQuality,
} from '../src/playback/metric-sample.ts';

const id = '12345678-1234-4123-8123-123456789abc';

test('uses UUID v4 without randomUUID and normalizes supported quality choices', () => {
  const original = Object.getOwnPropertyDescriptor(crypto, 'randomUUID');
  Object.defineProperty(crypto, 'randomUUID', { configurable: true, value: undefined });
  try {
    const ids = new Set(Array.from({ length: 32 }, playbackMetricID));
    assert.equal(ids.size, 32);
    for (const value of ids)
      assert.match(value, /^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$/);
  } finally {
    if (original) Object.defineProperty(crypto, 'randomUUID', original);
    else delete crypto.randomUUID;
  }
  for (const quality of ['auto', 'original', '1080', '720', '480'])
    assert.equal(playbackQuality(quality), quality);
  assert.equal(playbackQuality('invalid'), 'auto');
});

test('reports cumulative playback and stalls without counting startup, seeks or pauses', () => {
  const sample = new MetricSample(id, 'transcode', '720', 100, true);
  sample.waiting(200);
  sample.playing(500);
  sample.waiting(1500);
  sample.waiting(1600);
  const interim = sample.snapshot(1750);
  assert.equal(interim.stalls, 1);
  assert.equal(interim.stall_ms, 250);
  assert.equal(interim.played_ms, 1000);
  sample.playing(2000);
  sample.beginSeek(2500);
  sample.waiting(2700);
  sample.endSeek(3200, false);
  sample.waiting(3400);
  sample.playing(3500);
  sample.stop(4000);
  sample.waiting(4200);
  assert.deepEqual(sample.snapshot(10000), {
    id,
    stream_id: id,
    burn_subtitle: true,
    mode: 'transcode',
    quality: '720',
    width: 0,
    height: 0,
    startup_ms: 400,
    stalls: 1,
    stall_ms: 500,
    played_ms: 2000,
    buffer_ms: 0,
  });
  assert.equal(interim.stall_ms, 250);
});

test('omits autoplay-blocked startup and excludes time in the back-forward cache', () => {
  const sample = new MetricSample(id, 'direct', 'original', 0);
  sample.blockAutoplay();
  sample.playing(5000);
  sample.suspend(6000);
  sample.playing(7000);
  sample.resume(11000, true);
  const snapshot = sample.snapshot(11500);
  assert.equal(snapshot.played_ms, 1500);
  assert.equal(snapshot.startup_ms, undefined);
  assert.equal(snapshot.stream_id, undefined);
});

test('reports contiguous buffered-ahead duration and optional frame counts', () => {
  const ranges = { length: 2, start: (i) => [0, 20][i], end: (i) => [8, 40][i] };
  assert.equal(bufferedAhead(ranges, 5), 3000);
  assert.equal(bufferedAhead(ranges, 10), 0);
  assert.equal(bufferedAhead(ranges, 35), 5000);
  const sample = new MetricSample(id, 'remux', 'auto', 0);
  sample.readVideo({ videoWidth: 1920, videoHeight: 1080 }, 0);
  assert.equal(sample.snapshot(0).total_frames, undefined);
  sample.readVideo(
    {
      videoWidth: 1920,
      videoHeight: 1080,
      getVideoPlaybackQuality: () => ({ droppedVideoFrames: 1, totalVideoFrames: 100 }),
    },
    5000,
  );
  const snapshot = sample.snapshot(0);
  assert.equal(snapshot.width, 1920);
  assert.equal(snapshot.height, 1080);
  assert.equal(snapshot.buffer_ms, 5000);
  assert.equal(snapshot.dropped_frames, 1);
  assert.equal(snapshot.total_frames, 100);
});
