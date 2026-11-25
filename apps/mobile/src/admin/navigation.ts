import { serverURL } from '@dinosaur/client';

export function adminNavigationAllowed(origin: string, url: string) {
  if (url === 'about:blank') return true;
  try {
    const target = new URL(url);
    return target.origin === origin && !target.username && !target.password;
  } catch {
    return false;
  }
}

export function adminBootstrap(origin: string, code: string) {
  const server = serverURL(origin);
  if (!/^[A-Za-z0-9_-]{43}$/.test(code)) throw new Error('管理登录信息无效，请重试。');
  return {
    baseUrl: server + '/',
    html: `<!doctype html><meta name="viewport" content="width=device-width,initial-scale=1">
<body style="background:#111113"><form method="post" action="${server}/api/devices/admin-web/exchange">
<input type="hidden" name="code" value="${code}"></form><script>
try { sessionStorage.setItem('dinosaur-mobile-admin','1'); } catch {}
document.forms[0].submit();</script>`,
  };
}
