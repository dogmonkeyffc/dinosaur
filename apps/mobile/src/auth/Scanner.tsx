import { useEffect, useRef, useState } from 'react';
import { Linking, Platform, Text, View } from 'react-native';
import { CameraView, useCameraPermissions } from 'expo-camera';
import { Client, loginQR } from '@dinosaur/client';
import type { Account } from '@dinosaur/contracts';
import {
  Button,
  ErrorText,
  Input,
  Screen,
  probeServer,
  saveAccount,
  styles,
  useTask,
} from '@dinosaur/native-kit';

export function Scanner({
  initial,
  onLogin,
  close,
}: {
  initial?: string;
  onLogin: (account: Account) => void;
  close: () => void;
}) {
  const [permission, requestPermission] = useCameraPermissions();
  const [raw, setRaw] = useState(initial || '');
  const [scan, setScan] = useState<{ server: string; ticket: string }>();
  const [error, setError] = useState('');
  const [name, setName] = useState(Platform.OS === 'ios' ? '我的 iPhone / iPad' : '我的 Android 手机');
  const locked = useRef(false);
  const active = useRef(true);
  const task = useTask();
  useEffect(() => {
    active.current = true;
    return () => {
      active.current = false;
    };
  }, []);
  function accept(data: string) {
    if (locked.current) return;
    try {
      const qr = loginQR(data);
      locked.current = true;
      setScan(qr);
      setError('');
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : '二维码无效。');
    }
  }
  useEffect(() => {
    if (initial) accept(initial);
  }, [initial]);
  async function exchange() {
    if (!scan) return;
    const server = await probeServer(scan.server);
    if (!active.current) return;
    const client = new Client(server.url);
    try {
      const session = await client.exchange(scan.ticket, name.trim());
      const account = await saveAccount(server, session);
      if (active.current) onLogin(account);
    } finally {
      client.dispose();
    }
  }
  return (
    <Screen title="扫码登录">
      <Button onPress={close}>返回</Button>
      {scan ? (
        <View style={styles.card}>
          <Text style={styles.subtitle}>登录到此服务器</Text>
          <Text selectable style={styles.text}>
            {scan.server}
          </Text>
          {scan.server.startsWith('http:') && <Text style={styles.muted}>HTTP · 局域网连接</Text>}
          <Input
            accessibilityLabel="设备名称"
            value={name}
            onChangeText={setName}
            maxLength={48}
            placeholder="设备名称"
          />
          <Button primary disabled={task.busy || !name.trim()} onPress={() => task.run(exchange)}>
            {task.busy ? '正在登录…' : '确认登录'}
          </Button>
          <Button
            disabled={task.busy}
            onPress={() => {
              locked.current = false;
              setScan(undefined);
              task.clear();
            }}
          >
            重新扫码
          </Button>
        </View>
      ) : (
        <>
          {permission?.granted ? (
            <CameraView
              style={{ height: 300, borderRadius: 16 }}
              facing="back"
              barcodeScannerSettings={{ barcodeTypes: ['qr'] }}
              onBarcodeScanned={({ data }) => accept(data)}
            />
          ) : (
            <Button
              onPress={() => {
                if (permission?.canAskAgain === false) Linking.openSettings();
                else requestPermission();
              }}
            >
              允许使用相机
            </Button>
          )}
          <Input
            accessibilityLabel="登录链接"
            placeholder="或粘贴 Web 登录二维码的链接"
            value={raw}
            onChangeText={setRaw}
          />
          <Button disabled={!raw.trim()} onPress={() => accept(raw.trim())}>
            读取登录链接
          </Button>
        </>
      )}
      <ErrorText error={error || task.error} />
    </Screen>
  );
}
