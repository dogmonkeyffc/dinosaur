import { useState } from 'react';
import { Text, View } from 'react-native';
import type { ConnectedServer } from '@dinosaur/contracts';
import { Button, ErrorText, Input, Screen, probeServer, styles, useTask } from '@dinosaur/native-kit';
import { useDiscovery } from './useDiscovery';

export function Servers({ select, error }: { select: (server: ConnectedServer) => void; error?: string }) {
  const discovery = useDiscovery();
  const [address, setAddress] = useState('');
  const task = useTask();
  return (
    <Screen title="Dinosaur TV" tv>
      <Text style={styles.subtitle}>选择服务器</Text>
      {discovery.servers.map((server, index) => (
        <View key={server.id} style={styles.card}>
          <Text style={styles.subtitle}>{server.name}</Text>
          <Text style={styles.muted}>{server.url}</Text>
          <Button primary preferred={index === 0} onPress={() => select(server)}>
            连接服务器
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
      <View style={styles.card}>
        <Text style={styles.subtitle}>手动连接</Text>
        <Input
          accessibilityLabel="服务器地址"
          placeholder="http://192.168.1.100:8096"
          value={address}
          onChangeText={setAddress}
          keyboardType="url"
        />
        <Button
          disabled={task.busy || !address.trim()}
          onPress={() => task.run(async () => select(await probeServer(address)))}
        >
          连接此地址
        </Button>
      </View>
      <ErrorText error={error || task.error || discovery.error} />
    </Screen>
  );
}
