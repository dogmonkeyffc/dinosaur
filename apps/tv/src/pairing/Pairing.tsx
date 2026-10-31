import { useEffect, useState } from 'react';
import { Text, View } from 'react-native';
import { Client } from '@dinosaur/client';
import type { Account, ConnectedServer, TVPairing } from '@dinosaur/contracts';
import { Button, ErrorText, Screen, saveSingleAccount, styles } from '@dinosaur/native-kit';

export function Pairing({
  server,
  accepted,
  cancel,
}: {
  server: ConnectedServer;
  accepted: (account: Account) => void;
  cancel: () => void;
}) {
  const [pairing, setPairing] = useState<TVPairing>();
  const [now, setNow] = useState(Date.now());
  const [error, setError] = useState('');
  const [revision, setRevision] = useState(0);
  useEffect(() => {
    const client = new Client(server.url);
    let active = true;
    let secret: string | undefined;
    let completed = false;
    let timer: ReturnType<typeof setTimeout>;
    setPairing(undefined);
    setError('');
    async function begin() {
      try {
        const result = await client.createPairing('Dinosaur TV');
        secret = result.device_code;
        if (!active) return;
        setNow(Date.now());
        setPairing(result);
        async function poll() {
          try {
            const session = await client.poll(result.device_code);
            if (!active) return;
            if ('token' in session) {
              completed = true;
              const account = await saveSingleAccount(server, session);
              if (active) accepted(account);
              return;
            }
            timer = setTimeout(poll, Math.max(3, session.interval) * 1000);
          } catch (reason) {
            if (active) setError(reason instanceof Error ? reason.message : '配对失败。');
          }
        }
        timer = setTimeout(poll, Math.max(3, result.interval) * 1000);
      } catch (reason) {
        if (active) setError(reason instanceof Error ? reason.message : '无法创建配对码。');
      }
    }
    begin();
    const clock = setInterval(() => {
      if (active) setNow(Date.now());
    }, 1000);
    return () => {
      active = false;
      clearTimeout(timer);
      clearInterval(clock);
      client.dispose();
      if (secret && !completed) {
        const cleanup = new Client(server.url);
        cleanup
          .cancelPairing(secret)
          .catch(() => {})
          .finally(() => cleanup.dispose());
      }
    };
  }, [server.id, server.url, revision]);
  const remaining = pairing ? Math.max(0, Math.ceil(pairing.expires_at - now / 1000)) : 0;
  return (
    <Screen title="登录电视" tv>
      <Text style={styles.subtitle}>{server.name}</Text>
      <Text style={styles.muted}>{server.url}</Text>
      <View style={[styles.card, { alignItems: 'center', padding: 36 }]}>
        <Text style={styles.text}>手机 App → 登录电视</Text>
        <Text selectable style={{ color: '#c2df9e', fontSize: 76, fontWeight: '700', letterSpacing: 14 }}>
          {pairing && remaining > 0 ? pairing.code : pairing ? '已过期' : '••••••'}
        </Text>
        {pairing && remaining > 0 && <Text style={styles.muted}>{remaining} 秒后过期</Text>}
      </View>
      <ErrorText error={error} />
      <View style={styles.row}>
        <Button preferred onPress={cancel}>
          返回服务器列表
        </Button>
        <Button onPress={() => setRevision((value) => value + 1)}>刷新配对码</Button>
      </View>
    </Screen>
  );
}
