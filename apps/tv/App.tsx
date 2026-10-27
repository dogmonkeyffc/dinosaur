import { useEffect, useState } from 'react';
import { BackHandler, Text, View } from 'react-native';
import { ApiError, Client } from '@dinosaur/client';
import type { Account, ConnectedServer } from '@dinosaur/contracts';
import { Busy, ErrorText, loadAccounts, removeAccount, styles, useTask } from '@dinosaur/native-kit';
import { Servers } from './src/discovery/Servers';
import { Pairing } from './src/pairing/Pairing';
import { Home } from './src/catalog/Home';

export default function App() {
  const [account, setAccount] = useState<Account>();
  const [server, setServer] = useState<ConnectedServer>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const task = useTask();
  useEffect(() => {
    let active = true;
    let client: Client | undefined;
    async function restore() {
      try {
        const saved = (await loadAccounts())[0];
        if (!saved || !active) return;
        client = new Client(saved.server.url, saved.token);
        try {
          const result = await client.me();
          if (active) setAccount({ ...saved, user: result.user });
        } catch (reason) {
          if (reason instanceof ApiError && reason.status === 401) await removeAccount(saved.key);
          if (active) setError('无法恢复登录，请重新选择服务器并配对。');
        }
      } catch {
        if (active) setError('无法读取设备账号。');
      } finally {
        if (active) setLoading(false);
        client?.dispose();
      }
    }
    restore();
    return () => {
      active = false;
      client?.dispose();
    };
  }, []);
  useEffect(() => {
    const handler = BackHandler.addEventListener('hardwareBackPress', () => {
      if (!account && server) {
        setServer(undefined);
        return true;
      }
      return false;
    });
    return () => handler.remove();
  }, [account, server]);
  if (loading)
    return (
      <View style={styles.centered}>
        <Busy />
        <Text style={styles.text}>正在连接服务器…</Text>
      </View>
    );
  if (account)
    return (
      <View style={styles.screen}>
        <Home
          key={account.token}
          account={account}
          logout={() =>
            task.run(async () => {
              const client = new Client(account.server.url, account.token);
              try {
                await client.logout();
              } catch {
                /* A revoked or unreachable session can still be removed locally. */
              } finally {
                client.dispose();
              }
              for (const saved of await loadAccounts()) await removeAccount(saved.key);
              setAccount(undefined);
              setServer(undefined);
            })
          }
        />
        <ErrorText error={task.error} />
      </View>
    );
  if (server)
    return (
      <Pairing
        server={server}
        cancel={() => setServer(undefined)}
        accepted={(next) => {
          setAccount(next);
          setError('');
        }}
      />
    );
  return <Servers select={setServer} error={error} />;
}
