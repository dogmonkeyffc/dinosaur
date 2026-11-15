import { Client, serverURL } from '@dinosaur/client';
import type { ConnectedServer } from '@dinosaur/contracts';

export async function probeServer(raw: string, signal?: AbortSignal): Promise<ConnectedServer> {
  const client = new Client(serverURL(raw));
  const controller = new AbortController();
  const cancel = () => {
    controller.abort();
    client.dispose();
  };
  signal?.addEventListener('abort', cancel, { once: true });
  if (signal?.aborted) cancel();
  const timeout = setTimeout(cancel, 4000);
  try {
    const info = await client.info(controller.signal);
    if (controller.signal.aborted) throw new Error('请求已取消。');
    if (
      !/^[a-f0-9]{32}$/.test(info.id) ||
      info.api_version !== 1 ||
      typeof info.name !== 'string' ||
      !info.name.trim() ||
      !Array.isArray(info.capabilities) ||
      !info.capabilities.includes('tv_pairing')
    ) {
      throw new Error('此服务器尚未支持 App 登录，请先升级服务器。');
    }
    return { ...info, url: client.origin };
  } finally {
    clearTimeout(timeout);
    signal?.removeEventListener('abort', cancel);
    client.dispose();
  }
}
