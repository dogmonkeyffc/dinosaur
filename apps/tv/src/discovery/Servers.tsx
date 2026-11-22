import { useState } from 'react';
import { Text, View } from 'react-native';
import type { ConnectedServer } from '@dinosaur/contracts';
import { Button, ErrorText, Input, probeServer, styles, useTask } from '@dinosaur/native-kit';
import { Entry } from '../layout/Entry';
import { useDiscovery } from './useDiscovery';

export function Servers({ select, error }: { select: (server: ConnectedServer) => void; error?: string }) {
  const discovery = useDiscovery();
  const [address, setAddress] = useState('');
  const task = useTask();
  return (
    <Entry title="选择服务器">
      {discovery.servers.map((server, index) => (
        <View key={server.id} style={styles.card}>
          <Button primary preferred={index === 0} onPress={() => select(server)}>
            {server.name} · {server.url}
          </Button>
        </View>
      ))}
      {!discovery.servers.length && (
        <Text style={styles.muted}>{discovery.scanning ? '正在查找局域网内的服务器…' : '未发现服务器'}</Text>
      )}
      <View style={styles.row}>
        <Button preferred={!discovery.servers.length} onPress={discovery.restart}>
          重新扫描
        </Button>
      </View>
      <Input
        accessibilityLabel="服务器地址"
        placeholder="服务器地址"
        value={address}
        onChangeText={setAddress}
        keyboardType="url"
      />
      <Button
        primary
        disabled={task.busy || !address.trim()}
        onPress={() => task.run(async () => select(await probeServer(address)))}
      >
        连接此地址
      </Button>
      <ErrorText error={error || task.error || discovery.error} />
    </Entry>
  );
}
