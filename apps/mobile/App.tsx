import { useEffect, useRef, useState } from 'react';
import { Alert, BackHandler, Linking, View } from 'react-native';
import { Client } from '@dinosaur/client';
import type { Account } from '@dinosaur/contracts';
import { Busy, ErrorText, loadAccounts, removeAccount, styles, useTask } from '@dinosaur/native-kit';
import { Scanner } from './src/auth/Scanner';
import { Accounts } from './src/auth/Accounts';
import { PairTV } from './src/pairing/PairTV';
import { Home } from './src/catalog/Home';

type Route = 'accounts' | 'scan' | 'home' | 'pair';
export default function App() {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [active, setActive] = useState('');
  const [route, setRoute] = useState<Route>('accounts');
  const [returnTo, setReturnTo] = useState<Route>('home');
  const [initialQR, setInitialQR] = useState<string>();
  const [loading, setLoading] = useState(true);
  const task = useTask();
  const incomingLink = useRef(false);
  useEffect(() => {
    let alive = true;
    loadAccounts()
      .then((items) => {
        if (alive) {
          setAccounts(items);
          if (items[0]) {
            setActive(items[0].key);
            if (!incomingLink.current) setRoute('home');
          }
        }
      })
      .catch(() => {
        if (alive) Alert.alert('无法读取账号', '请检查系统安全存储后重试。');
      })
      .finally(() => {
        if (alive) setLoading(false);
      });
    const accept = (url: string | null | undefined) => {
      if (alive && url?.startsWith('dinosaur://login')) {
        incomingLink.current = true;
        setInitialQR(url);
        setRoute('scan');
      }
    };
    Linking.getInitialURL().then(accept);
    const listener = Linking.addEventListener('url', ({ url }) => accept(url));
    return () => {
      alive = false;
      listener.remove();
    };
  }, []);
  useEffect(() => {
    const listener = BackHandler.addEventListener('hardwareBackPress', () => {
      if (route === 'home' || route === 'accounts') return false;
      setInitialQR(undefined);
      setRoute(accounts.length ? 'home' : 'accounts');
      return true;
    });
    return () => listener.remove();
  }, [route, accounts.length]);
  function choose(account: Account) {
    setActive(account.key);
    setRoute(returnTo);
  }
  function scan() {
    setInitialQR(undefined);
    setRoute('scan');
  }
  function picker(next: Route) {
    setReturnTo(next);
    setRoute('accounts');
  }
  async function logout(account: Account) {
    const client = new Client(account.server.url, account.token);
    try {
      await client.logout();
    } catch {
      /* Local sign-out remains available while the server is offline. */
    } finally {
      client.dispose();
    }
    await removeAccount(account.key);
    setAccounts((items) => items.filter((item) => item.key !== account.key));
    if (active === account.key) setActive('');
  }
  if (loading)
    return (
      <View style={styles.centered}>
        <Busy />
      </View>
    );
  const account = accounts.find((item) => item.key === active);
  if (route === 'scan')
    return (
      <Scanner
        key={initialQR || 'camera'}
        initial={initialQR}
        close={() => setRoute(account ? 'home' : 'accounts')}
        onLogin={(item) => {
          setAccounts((items) => [item, ...items.filter((old) => old.key !== item.key)]);
          setActive(item.key);
          setInitialQR(undefined);
          setRoute('home');
        }}
      />
    );
  if (route === 'pair' && account)
    return (
      <PairTV
        key={account.key}
        account={account}
        switchAccount={() => picker('pair')}
        close={() => setRoute('home')}
      />
    );
  if (route === 'home' && account)
    return (
      <Home
        key={account.key}
        account={account}
        pair={() => setRoute('pair')}
        accounts={() => picker('home')}
      />
    );
  return (
    <View style={styles.screen}>
      <Accounts
        accounts={accounts}
        select={choose}
        scan={scan}
        logout={(item) => task.run(() => logout(item))}
        busy={task.busy}
        close={account ? () => setRoute('home') : undefined}
      />
      <ErrorText error={task.error} />
    </View>
  );
}
