import { useEffect, useRef, useState } from 'react';
import { Platform, Text, View } from 'react-native';
import { Client } from '@dinosaur/client';
import type { Account, ConnectedServer } from '@dinosaur/contracts';
import { Button, ErrorText, Input, Screen, saveAccount, styles, useTask } from '@dinosaur/native-kit';
import { ServerPicker } from './ServerPicker';

export function CodeLogin({ onLogin, close }: {
  onLogin: (account: Account) => void;
  close: () => void;
}) {
  const [server, setServer] = useState<ConnectedServer>();
  if (!server) return <ServerPicker select={setServer} close={close} />;
  return (
    <CodeExchange
      key={server.url}
      server={server}
      onLogin={onLogin}
      close={close}
      changeServer={() => setServer(undefined)}
    />
  );
}

function CodeExchange({ server, onLogin, close, changeServer }: {
  server: ConnectedServer;
  onLogin: (account: Account) => void;
  close: () => void;
  changeServer: () => void;
}) {
  const [code, setCode] = useState('');
  const [name, setName] = useState(Platform.OS === 'ios' ? '我的 iPhone / iPad' : '我的 Android 手机');
  const active = useRef(false);
  const client = useRef<Client | undefined>(undefined);
  const task = useTask();
  useEffect(() => {
    active.current = true;
    return () => {
      active.current = false;
      client.current?.dispose();
    };
  }, []);
  function leave(next: () => void) {
    active.current = false;
    client.current?.dispose();
    setCode('');
    next();
  }
  async function exchange() {
    if (!/^\d{8}$/.test(code) || !name.trim()) return;
    const connection = new Client(server.url);
    client.current = connection;
    try {
      const session = await connection.exchangeLoginCode(code, name.trim());
      if (!active.current) return;
      const account = await saveAccount(server, session);
      if (active.current) {
        setCode('');
        onLogin(account);
      }
    } finally {
      connection.dispose();
      if (client.current === connection) client.current = undefined;
    }
  }
  return (
    <Screen title="配对码登录">
      <Button onPress={() => leave(close)}>返回</Button>
      <View style={styles.card}>
        <Text style={styles.subtitle}>{server.name}</Text>
        <Text selectable style={styles.text}>{server.url}</Text>
        <Text style={styles.muted}>
          {server.url.startsWith('http:') ? 'HTTP · 未加密连接' : 'HTTPS · 加密连接'}
        </Text>
        <Button onPress={() => leave(changeServer)}>更换服务器</Button>
      </View>
      <Input
        accessibilityLabel="Web 账户安全页的 8 位配对码"
        placeholder="输入 Web 账户安全页配对码"
        value={code}
        onChangeText={setCode}
        keyboardType="number-pad"
        maxLength={8}
        autoCapitalize="none"
        autoCorrect={false}
        secureTextEntry
        editable={!task.busy}
      />
      <Input
        accessibilityLabel="设备名称"
        placeholder="设备名称"
        value={name}
        onChangeText={setName}
        maxLength={48}
        editable={!task.busy}
      />
      <Text style={styles.text}>授权账号以 Web 显示为准</Text>
      <Button
        primary
        disabled={task.busy || !/^\d{8}$/.test(code) || !name.trim()}
        onPress={() => task.run(exchange)}
      >
        {task.busy ? '正在登录…' : '确认此服务器并登录'}
      </Button>
      <ErrorText error={task.error} />
    </Screen>
  );
}
