import { Text, View } from 'react-native';
import type { Account } from '@dinosaur/contracts';
import { Button, Screen, styles } from '@dinosaur/native-kit';

export function Accounts({
  accounts,
  select,
  scan,
  logout,
  busy,
  close,
}: {
  accounts: Account[];
  select: (account: Account) => void;
  scan: () => void;
  logout: (account: Account) => void;
  busy: boolean;
  close?: () => void;
}) {
  return (
    <Screen title="Dinosaur">
      {close && <Button onPress={close}>返回</Button>}
      {accounts.map((account) => (
        <View style={styles.card} key={account.key}>
          <Text style={styles.subtitle}>{account.user.username}</Text>
          <Text style={styles.muted}>
            {account.server.name} · {account.server.url}
          </Text>
          <View style={styles.row}>
            <Button primary disabled={busy} onPress={() => select(account)}>
              使用此账号
            </Button>
            <Button disabled={busy} onPress={() => logout(account)}>
              退出账号
            </Button>
          </View>
        </View>
      ))}
      <Button primary onPress={scan}>
        扫描 Web 登录二维码
      </Button>
    </Screen>
  );
}
