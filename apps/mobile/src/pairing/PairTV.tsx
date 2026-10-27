import { useState } from 'react';
import { Text, View } from 'react-native';
import type { Account, Pairing } from '@dinosaur/contracts';
import { Button, ErrorText, Input, Screen, styles, useClient, useTask } from '@dinosaur/native-kit';

export function PairTV({
  account,
  switchAccount,
  close,
}: {
  account: Account;
  switchAccount: () => void;
  close: () => void;
}) {
  const client = useClient(account);
  const task = useTask();
  const [code, setCode] = useState('');
  const [pairing, setPairing] = useState<Pairing>();
  const [done, setDone] = useState(false);
  return (
    <Screen title="登录电视">
      <View style={styles.row}>
        <Button onPress={close}>返回</Button>
        <Button onPress={switchAccount}>切换账号</Button>
      </View>
      <View style={styles.card}>
        <Text style={styles.subtitle}>登录账号：{account.user.username}</Text>
        <Text style={styles.muted}>
          {account.server.name} · {account.server.url}
        </Text>
      </View>
      {done ? (
        <View style={styles.card}>
          <Text style={styles.subtitle}>已授权 {pairing?.name}</Text>
          <Text style={styles.text}>电视将以 {account.user.username} 登录</Text>
          <Button primary onPress={close}>
            完成
          </Button>
        </View>
      ) : pairing ? (
        <View style={styles.card}>
          <Text style={styles.subtitle}>{pairing.name}</Text>
          <Text style={styles.text}>使用 {account.user.username} 登录此电视？</Text>
          <Button
            primary
            disabled={task.busy}
            onPress={() =>
              task.run(async () => {
                await client.approve(pairing.pairing_id);
                setDone(true);
              })
            }
          >
            确认授权
          </Button>
          <Button
            disabled={task.busy}
            onPress={() => {
              setPairing(undefined);
              task.clear();
            }}
          >
            重新输入
          </Button>
        </View>
      ) : (
        <View style={styles.card}>
          <Input
            accessibilityLabel="电视六位配对码"
            placeholder="输入电视上的六位配对码"
            keyboardType="number-pad"
            maxLength={6}
            value={code}
            onChangeText={(value) => setCode(value.replace(/\D/g, ''))}
          />
          <Button
            primary
            disabled={task.busy || code.length !== 6}
            onPress={() =>
              task.run(async () => {
                setPairing(await client.lookup(code));
              })
            }
          >
            查找电视
          </Button>
        </View>
      )}
      <ErrorText error={task.error} />
    </Screen>
  );
}
