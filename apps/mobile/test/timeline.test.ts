import assert from 'node:assert/strict';
import test from 'node:test';
import type { WatchHistoryEntry } from '@dinosaur/contracts';
import { mergeTimeline, timelineSections } from '../src/history/timeline';

function item(id: number, watched_on: string): WatchHistoryEntry {
  return {
    id,
    watched_on,
    title: '影片',
    original_title: '',
    year: 2024,
    rating: 5,
    review: '',
    source: 'douban',
    source_id: String(id),
  };
}

test('timeline groups calendar dates by month without timezone shifts', () => {
  const sections = timelineSections([item(1, '2024-01-01'), item(3, '2025-01-01'), item(2, '2024-01-31')]);
  assert.deepEqual(
    sections.map((section) => section.title),
    ['2025年1月', '2024年1月'],
  );
  assert.deepEqual(
    sections[1].data.map((entry) => entry.id),
    [2, 1],
  );
  assert.deepEqual(timelineSections([]), []);
});

test('overlapping pages do not duplicate timeline records and preserve edits', () => {
  const entries = mergeTimeline(
    [item(1, '2024-01-01')],
    [{ ...item(1, '2024-01-01'), review: '编辑' }, item(2, '2023-01-01')],
  );
  assert.equal(entries.length, 2);
  assert.equal(entries[0].review, '编辑');
});
