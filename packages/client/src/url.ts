export function serverURL(raw: string): string {
  const url = new URL(raw.trim());
  if (
    !['http:', 'https:'].includes(url.protocol) ||
    url.username ||
    url.password ||
    url.search ||
    url.hash ||
    !['', '/'].includes(url.pathname)
  ) {
    throw new Error('服务器地址必须是完整的 HTTP(S) 地址，不包含路径或凭据。');
  }
  if (url.origin.length > 512) throw new Error('服务器地址过长。');
  if (url.protocol === 'http:' && !localHost(url.hostname)) {
    throw new Error('HTTP 仅用于局域网 IP 或 .local 地址；其他地址请使用 HTTPS。');
  }
  return url.origin;
}

function localHost(host: string): boolean {
  if (host === 'localhost' || host.endsWith('.local')) return true;
  const parts = host.split('.').map(Number);
  if (parts.length !== 4 || parts.some((x) => !Number.isInteger(x) || x < 0 || x > 255)) return false;
  return (
    parts[0] === 10 ||
    parts[0] === 127 ||
    (parts[0] === 192 && parts[1] === 168) ||
    (parts[0] === 172 && parts[1] >= 16 && parts[1] <= 31)
  );
}

export function loginQR(raw: string): { server: string; ticket: string } {
  if (raw.length > 4096) throw new Error('二维码内容过长。');
  const url = new URL(raw);
  if (
    url.protocol !== 'dinosaur:' ||
    url.hostname !== 'login' ||
    (url.pathname && url.pathname !== '/') ||
    url.hash ||
    url.username ||
    url.password ||
    url.port ||
    url.searchParams.getAll('server').length !== 1 ||
    url.searchParams.getAll('ticket').length !== 1
  ) {
    throw new Error('这不是 Dinosaur 登录二维码。');
  }
  const ticket = url.searchParams.get('ticket') || '';
  if (!/^[A-Za-z0-9_-]{43}$/.test(ticket)) throw new Error('二维码无效。');
  return { server: serverURL(url.searchParams.get('server') || ''), ticket };
}

export function resourceURL(server: string, path: string): string {
  const base = serverURL(server);
  const url = new URL(path, base);
  if (url.origin !== base || !url.pathname.startsWith('/api/') || url.username || url.password) {
    throw new Error('服务器返回了不受信任的资源地址。');
  }
  return url.href;
}
