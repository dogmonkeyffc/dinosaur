import * as SecureStore from 'expo-secure-store';
import type { Account, ConnectedServer, DeviceSession } from '@dinosaur/contracts';
import { Client, serverURL } from '@dinosaur/client';

const indexKey = 'dinosaur.accounts';
const keyFor = (id: string) => 'dinosaur.account.' + id;
let queue: Promise<unknown> = Promise.resolve();

function serial<T>(operation: () => Promise<T>): Promise<T> {
  const result = queue.catch(() => {}).then(operation);
  queue = result;
  return result;
}
async function keys(): Promise<string[]> {
  const stored = await SecureStore.getItemAsync(indexKey);
  if (!stored) return [];
  const parsed: unknown = JSON.parse(stored);
  if (!Array.isArray(parsed)) return [];
  return parsed
    .filter((key): key is string => typeof key === 'string' && /^[a-f0-9]{32}_[0-9]+$/.test(key))
    .slice(0, 12);
}
export function loadAccounts(): Promise<Account[]> {
  return serial(async () => {
    const result: Account[] = [];
    for (const key of await keys()) {
      const stored = await SecureStore.getItemAsync(keyFor(key));
      if (!stored) continue;
      try {
        const account = JSON.parse(stored) as Account;
        if (
          account.key !== key ||
          !/^[A-Za-z0-9_-]{43}$/.test(account.token) ||
          !account.user?.id ||
          !account.user.username ||
          account.expires_at <= Date.now() / 1000
        )
          continue;
        account.server.url = serverURL(account.server.url);
        result.push(account);
      } catch {
        /* Ignore an invalid saved account; never log its credentials. */
      }
    }
    const valid = new Set(result.map((account) => account.key));
    for (const key of await keys()) {
      if (!valid.has(key)) await SecureStore.deleteItemAsync(keyFor(key));
    }
    await SecureStore.setItemAsync(indexKey, JSON.stringify([...valid]));
    return result;
  });
}
export function saveAccount(server: ConnectedServer, session: DeviceSession): Promise<Account> {
  return serial(async () => {
    if (
      !/^[a-f0-9]{32}$/.test(server.id) ||
      !/^[A-Za-z0-9_-]{43}$/.test(session.token) ||
      !Number.isSafeInteger(session.user.id) ||
      session.user.id < 1
    )
      throw new Error('服务器返回了无效的账号信息。');
    const key = server.id + '_' + session.user.id;
    const existing = await keys();
    const next = [key, ...existing.filter((x) => x !== key)];
    if (next.length > 12) throw new Error('最多保存 12 个账号，请先退出一个账号。');
    const account: Account = { ...session, key, server: { ...server, url: serverURL(server.url) } };
    const previous = await SecureStore.getItemAsync(keyFor(key));
    await SecureStore.setItemAsync(keyFor(key), JSON.stringify(account));
    await SecureStore.setItemAsync(indexKey, JSON.stringify(next));
    if (previous) {
      try {
        const old = JSON.parse(previous) as Account;
        if (old.token !== session.token && /^[A-Za-z0-9_-]{43}$/.test(old.token)) {
          const client = new Client(server.url, old.token);
          client
            .logout()
            .catch(() => {})
            .finally(() => client.dispose());
        }
      } catch {
        /* The new session is already safely stored. */
      }
    }
    return account;
  });
}
export function removeAccount(key: string): Promise<void> {
  return serial(async () => {
    await SecureStore.setItemAsync(indexKey, JSON.stringify((await keys()).filter((x) => x !== key)));
    await SecureStore.deleteItemAsync(keyFor(key));
  });
}

export async function saveSingleAccount(server: ConnectedServer, session: DeviceSession): Promise<Account> {
  const account = await saveAccount(server, session);
  for (const previous of await loadAccounts()) {
    if (previous.key === account.key) continue;
    await removeAccount(previous.key);
    const client = new Client(previous.server.url, previous.token);
    client
      .logout()
      .catch(() => {})
      .finally(() => client.dispose());
  }
  return account;
}
