import assert from 'node:assert/strict';
import test from 'node:test';
import { ApiError, Client, loginQR, resourceURL, serverURL } from '../src/index';

const token = 'a'.repeat(43);
test('QR and server validation reject credential forwarding to arbitrary origins', () => {
  assert.equal(serverURL(' http://192.168.31.197:18096/ '), 'http://192.168.31.197:18096');
  assert.deepEqual(
    loginQR(`dinosaur://login?server=${encodeURIComponent('https://dinosaur.example.com')}&ticket=${token}`),
    { server: 'https://dinosaur.example.com', ticket: token },
  );
  for (const url of [
    'javascript:alert(1)',
    'https://user:pass@example.com',
    'https://example.com/path',
    'https://example.com?token=secret',
    'http://example.com',
    'http://8.8.8.8',
  ])
    assert.throws(() => serverURL(url));
  for (const path of [
    'https://attacker.example/api/poster',
    '//attacker.example/api/poster',
    '/not-an-api',
    '/api/../login',
  ])
    assert.throws(() => resourceURL('https://dinosaur.example.com', path));
  assert.throws(() =>
    loginQR(`dinosaur://login?server=https://a.example&server=https://b.example&ticket=${token}`),
  );
  assert.throws(() => loginQR(`https://login?server=https://a.example&ticket=${token}`));
});

test('native API authentication stays in headers with no browser cookie', async () => {
  const calls: { url: string; options: RequestInit }[] = [];
  const fetcher = (async (url, options) => {
    calls.push({ url: String(url), options: options || {} });
    return new Response(JSON.stringify({ ok: true }), { status: 200 });
  }) as typeof fetch;
  const client = new Client('https://dinosaur.example.com', token, fetcher);
  await client.approve('pairing-id');
  assert.equal(calls[0].url, 'https://dinosaur.example.com/api/devices/pairings/approve');
  assert.equal(calls[0].options.credentials, 'omit');
  assert.equal(calls[0].options.redirect, 'error');
  assert.equal((calls[0].options.headers as Record<string, string>).Authorization, `Bearer ${token}`);
  assert.deepEqual(JSON.parse(String(calls[0].options.body)), { pairing_id: 'pairing-id' });
  assert.ok(!calls[0].url.includes(token));
  assert.equal(client.resource('/api/media/1/poster').headers.Authorization, `Bearer ${token}`);
  client.dispose();
  await assert.rejects(client.me(), /连接已关闭/);
});

test('disposing a session rejects a late response body', async () => {
  let release!: (value: unknown) => void;
  let reading!: () => void;
  const bodyRead = new Promise<void>((resolve) => {
    reading = resolve;
  });
  const fetcher = (async () => ({
    ok: true,
    json: () => {
      reading();
      return new Promise((resolve) => {
        release = resolve;
      });
    },
  })) as unknown as typeof fetch;
  const client = new Client('https://dinosaur.example.com', token, fetcher);
  const pending = client.me();
  await bodyRead;
  client.dispose();
  release({ user: { id: 1 } });
  await assert.rejects(pending, /请求已取消/);
});

test('pairing errors are actionable and polling can return pending', async () => {
  const pendingClient = new Client(
    'https://dinosaur.example.com',
    undefined,
    (async () =>
      new Response(JSON.stringify({ status: 'pending', interval: 3 }), { status: 202 })) as typeof fetch,
  );
  assert.deepEqual(await pendingClient.poll(token), { status: 'pending', interval: 3 });
  const expired = new Client(
    'https://dinosaur.example.com',
    undefined,
    (async () =>
      new Response(JSON.stringify({ error: { code: 'pairing_expired' } }), { status: 410 })) as typeof fetch,
  );
  await assert.rejects(
    expired.poll(token),
    (reason: unknown) =>
      reason instanceof ApiError && reason.status === 410 && reason.code === 'pairing_expired',
  );
});
