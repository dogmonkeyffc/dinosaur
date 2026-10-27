import { useEffect, useMemo, useRef, useState } from 'react';
import type { Account, ConnectedServer } from '@dinosaur/contracts';
import { Client } from '@dinosaur/client';

export function useClient(account: Account): Client {
  const client = useMemo(
    () => new Client(account.server.url, account.token),
    [account.server.url, account.token],
  );
  useEffect(() => () => client.dispose(), [client]);
  return client;
}
export function useTask() {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const active = useRef(true);
  const pending = useRef(false);
  useEffect(() => {
    active.current = true;
    return () => {
      active.current = false;
    };
  }, []);
  async function run(operation: () => Promise<void>) {
    if (pending.current) return;
    pending.current = true;
    setBusy(true);
    setError('');
    try {
      await operation();
    } catch (reason) {
      if (active.current) setError(reason instanceof Error ? reason.message : '操作失败。');
    } finally {
      pending.current = false;
      if (active.current) setBusy(false);
    }
  }
  return { busy, error, run, clear: () => setError('') };
}
export async function probeServer(url: string): Promise<ConnectedServer> {
  const client = new Client(url);
  const abort = new AbortController();
  const timeout = setTimeout(() => abort.abort(), 4000);
  try {
    const info = await client.info(abort.signal);
    if (
      !/^[a-f0-9]{32}$/.test(info.id) ||
      info.api_version !== 1 ||
      typeof info.name !== 'string' ||
      !Array.isArray(info.capabilities) ||
      !info.capabilities.includes('tv_pairing')
    )
      throw new Error('此服务器尚未支持 App 登录，请先升级服务器。');
    return { ...info, url: client.origin };
  } finally {
    clearTimeout(timeout);
    client.dispose();
  }
}
