import assert from 'node:assert/strict';
import test from 'node:test';
import { Client } from '../src/index';

const token = 'a'.repeat(43);
test('work catalog and member details preserve separate identities and authenticated requests', async () => {
  const urls: string[] = [];
  const work = { id: 8, title: '鬼灭之刃', library_id: 4, member_count: 2, episode_count: 38, season_count: 2, watched_count: 1 };
  const members = [
    { id: 12, title: '鬼灭之刃', work_id: 8, episode_count: 26 },
    { id: 19, title: '刀匠村篇', work_id: 8, episode_count: 12 },
  ];
  const fetcher = (async (url, options) => {
    urls.push(String(url));
    assert.equal((options?.headers as Record<string, string>).Authorization, `Bearer ${token}`);
    assert.equal(options?.credentials, 'omit');
    const path = new URL(String(url)).pathname;
    const body = path === '/api/works/8' ? { work, members }
      : path === '/api/series/19' ? { series: members[1], work, seasons: [{ number: 1, episodes: [] }] }
      : { items: [{ ...members[0], title: work.title, member_count: 2 }], page: 1, page_size: 24, total: 1 };
    return new Response(JSON.stringify(body));
  }) as typeof fetch;
  const client = new Client('https://dinosaur.example.com', token, fetcher);
  const page = await client.series(1, '刀匠村篇');
  assert.equal(page.items[0].work_id, 8);
  const detail = await client.work(page.items[0].work_id!);
  assert.deepEqual(detail.members.map((item) => item.id), [12, 19]);
  const chapter = await client.detail(detail.members[1].id);
  assert.equal(chapter.work?.id, 8);
  assert.equal(chapter.seasons[0].number, 1);
  assert.equal(chapter.series.id, 19);
  assert(urls[0].includes(encodeURIComponent('刀匠村篇')));
  assert.equal(urls[1], 'https://dinosaur.example.com/api/works/8');
  client.dispose();
});

test('missing work returns a typed failure instead of entering its representative chapter', async () => {
  const client = new Client('https://dinosaur.example.com', token, (async () =>
    new Response(JSON.stringify({ error: { code: 'not_found' } }), { status: 404 })) as typeof fetch);
  await assert.rejects(client.work(999), (error: unknown) => error instanceof Error && 'status' in error && error.status === 404);
  client.dispose();
});
