import assert from 'node:assert/strict';
import test from 'node:test';
import type { MediaLibrary, Movie } from '@dinosaur/contracts';
import { homeLibraries, homeRecent, parseHiddenLibraries } from '../src/settings/homeVisibility';

const libraries: MediaLibrary[] = [
  { id: 1, name: '电影', type: 'movie' },
  { id: 2, name: '剧集', type: 'tv' },
];

test('saved visibility defaults to all libraries and rejects malformed settings', () => {
  assert.deepEqual(parseHiddenLibraries(null), []);
  assert.deepEqual(parseHiddenLibraries('[2,2,1]'), [2, 1]);
  for (const value of ['{}', '[0]', '[-1]', '["2"]', '[1.5]', '[null]', 'invalid']) {
    assert.throws(() => parseHiddenLibraries(value));
  }
});

test('hiding libraries affects the home projection without removing catalog libraries', () => {
  assert.deepEqual(
    homeLibraries(libraries, [2]).map((library) => library.id),
    [1],
  );
  assert.deepEqual(homeLibraries(libraries, [1, 2]), []);
  assert.deepEqual(homeLibraries(libraries, [999]), libraries);
  assert.equal(libraries.length, 2);
});

test('continue watching excludes hidden and removed libraries, including TV episodes', () => {
  const base: Movie = {
    id: 1,
    title: '示例',
    type: 'movie',
    duration: 60,
    position: 20,
    watched: false,
    year: 2026,
  };
  const recent = [
    { ...base, id: 1, library_id: 1 },
    { ...base, id: 2, library_id: 2, type: 'tv' as const },
    { ...base, id: 3, library_id: 999 },
    { ...base, id: 4 },
  ];
  assert.deepEqual(
    homeRecent(recent, libraries, [2]).map((item) => item.id),
    [1],
  );
  assert.deepEqual(homeRecent(recent, libraries, [1, 2]), []);
  assert.deepEqual(
    homeRecent(recent, libraries, []).map((item) => item.id),
    [1, 2, 4],
  );
});
