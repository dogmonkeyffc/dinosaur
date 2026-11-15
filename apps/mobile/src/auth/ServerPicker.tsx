import { useEffect, useRef, useState } from 'react';
import { Text, View } from 'react-native';
import type { ConnectedServer } from '@dinosaur/contracts';
import { Button, ErrorText, Input, Screen, styles, useTask } from '@dinosaur/native-kit';
import { probeServer } from '../../../../packages/native-kit/src/discovery/probeServer';
import { useDiscovery } from '../../../../packages/native-kit/src/discovery/useDiscovery';

export function ServerPicker({ select, close }: {
  select: (server: ConnectedServer) => void;
  close: () => void;
}) {
  const discovery = useDiscovery();
  const [address, setAddress] = useState('');
  const request = useRef<AbortController | undefined>(undefined);
  const task = useTask();
  useEffect(() => () => request.current?.abort(), []);
  function connect(url: string) {
    task.run(async () => {
      request.current?.abort();
      const controller = new AbortController();
      request.current = controller;
      const server = await probeServer(url, controller.signal);
      if (!controller.signal.aborted) select(server);
    });
  }
  return (
    <Screen title="选择服务器">
      <Button onPress={() => { request.current?.abort(); close(); }}>返回</Button>
      <View style={styles.card}>
        <Input
          accessibilityLabel="服务器地址"
          placeholder="http://10.0.2.2:8096"
          value={address}
          onChangeText={setAddress}
          keyboardType="url"
          autoCapitalize="none"
          autoCorrect={false}
          editable={!task.busy}
        />
        <Button primary disabled={task.busy || !address.trim()} onPress={() => connect(address)}>
          {task.busy ? '正在确认服务器…' : '连接此地址'}
        </Button>
      </View>
      {discovery.servers.map((server) => (
        <Button key={server.id} disabled={task.busy} onPress={() => connect(server.url)}>
          {server.name} · {server.url}
        </Button>
      ))}
      {!discovery.servers.length && (
        <Text style={styles.muted}>
          {discovery.scanning ? '正在查找局域网内的服务器…' : '未发现服务器'}
        </Text>
      )}
      <Button disabled={task.busy} onPress={discovery.restart}>重新扫描</Button>
      <ErrorText error={task.error || discovery.error} />
    </Screen>
  );
}
