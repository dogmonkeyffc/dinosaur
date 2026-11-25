import assert from 'node:assert/strict';
import test from 'node:test';
import { adminBootstrap, adminNavigationAllowed } from '../src/admin/navigation';

test('embedded admin navigation stays on its authenticated server', () => {
  const origin = 'https://dinosaur.example.com';
  assert.equal(adminNavigationAllowed(origin, origin + '/settings/scraping'), true);
  assert.equal(adminNavigationAllowed(origin, 'about:blank'), true);
  for (const url of [
    'https://other.example/settings',
    'http://dinosaur.example.com',
    'javascript:alert(1)',
    'file:///data/test',
    'https://admin@dinosaur.example.com',
  ]) {
    assert.equal(adminNavigationAllowed(origin, url), false);
  }
});

test('admin handoff posts a validated one-time code from its server origin', () => {
  const source = adminBootstrap('https://dinosaur.example.com', 'a'.repeat(43));
  assert.equal(source.baseUrl, 'https://dinosaur.example.com/');
  assert.match(source.html, /method="post"/);
  assert.match(source.html, /dinosaur-mobile-admin/);
  assert.throws(() => adminBootstrap('https://dinosaur.example.com', '"><script>bad</script>'));
  assert.throws(() => adminBootstrap('https://dinosaur.example.com', 'expired'));
});
