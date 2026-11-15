import assert from 'node:assert/strict';
import test from 'node:test';
import { ApiError, Client } from '../src/index';

test('login code is exchanged in JSON without placing credentials in URLs or cookies', async () => {
  const session = {
    token: 'a'.repeat(43),
    user: { id: 3, username: 'viewer', role: 'user' },
    expires_at: 999,
  };
  let called = false;
  const fetcher = (async (url, options) => {
    called = true;
    assert.equal(String(url), 'http://192.168.31.197:18096/api/devices/login-code/exchange');
    assert.equal(options?.method, 'POST');
    assert.equal(options?.credentials, 'omit');
    assert.equal(options?.redirect, 'error');
    assert.deepEqual(JSON.parse(String(options?.body)), { code: '00123456', name: 'Pixel AVD' });
    assert.equal((options?.headers as Record<string, string>).Authorization, undefined);
    return new Response(JSON.stringify(session));
  }) as typeof fetch;
  const client = new Client('http://192.168.31.197:18096', undefined, fetcher);
  assert.deepEqual(await client.exchangeLoginCode('00123456', 'Pixel AVD'), session);
  assert(called);
  client.dispose();
});

test('expired login code and rate limits retain actionable typed errors', async () => {
  for (const [status, code] of [
    [410, 'pairing_expired'],
    [429, 'rate_limited'],
  ] as const) {
    const client = new Client(
      'https://dinosaur.example.com',
      undefined,
      (async () => new Response(JSON.stringify({ error: { code } }), { status })) as typeof fetch,
    );
    await assert.rejects(
      client.exchangeLoginCode('12345678', 'Pixel'),
      (error: unknown) => error instanceof ApiError && error.code === code && error.status === status,
    );
    client.dispose();
  }
});
