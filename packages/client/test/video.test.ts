import assert from 'node:assert/strict';
import test from 'node:test';
import { Client } from '../src/index';

test('ordinary videos use their own catalog and continue-watching scope', async () => {
  const urls: URL[] = [];
  const client = new Client('https://dinosaur.example.com', 'a'.repeat(43), (async (url) => {
    urls.push(new URL(String(url)));
    return new Response(JSON.stringify({ items: [], total: 0 }));
  }) as typeof fetch);
  await client.movies(2, 'S01E01', 7, 'video');
  await client.recent('video');
  assert.equal(urls[0].searchParams.get('type'), 'video');
  assert.equal(urls[0].searchParams.get('library_id'), '7');
  assert.equal(urls[0].searchParams.get('page'), '2');
  assert.equal(urls[1].searchParams.get('type'), 'video');
});
