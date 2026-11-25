import { Pressable, Text, View } from 'react-native';
import type { Account } from '@dinosaur/contracts';
import { MobileButton, MobilePage } from '../ui/MobilePage';
import { pageStyles as css } from '../ui/pageStyles';

export function Accounts({
  accounts,
  activeKey,
  select,
  scan,
  codeLogin,
  logout,
  busy,
  close,
}: {
  accounts: Account[];
  activeKey?: string;
  select: (account: Account) => void;
  scan: () => void;
  codeLogin: () => void;
  logout: (account: Account) => void;
  busy: boolean;
  close?: () => void;
}) {
  return (
    <MobilePage title={accounts.length ? '切换账号' : '登录 Dinosaur'} close={close}>
      {accounts.map((account) => (
        <View style={css.card} key={account.key}>
          <Pressable
            accessibilityRole="button"
            accessibilityLabel={`使用账号 ${account.user.username}`}
            accessibilityState={{ selected: activeKey === account.key }}
            disabled={busy}
            onPress={() => select(account)}
            style={css.account}
          >
            <View style={css.avatar}>
              <Text style={css.initial}>{account.user.username.slice(0, 1).toUpperCase()}</Text>
            </View>
            <View style={css.identity}>
              <Text style={css.name}>{account.user.username}</Text>
              <Text style={css.server} numberOfLines={1}>
                {account.server.name || account.server.url}
              </Text>
            </View>
            {activeKey === account.key && <Text style={css.selection}>✓</Text>}
          </Pressable>
          <View style={css.actions}>
            <MobileButton disabled={busy} press={() => logout(account)}>
              退出账号
            </MobileButton>
          </View>
        </View>
      ))}
      <View style={css.card}>
        <MobileButton primary disabled={busy} press={codeLogin}>
          配对码登录
        </MobileButton>
        <MobileButton disabled={busy} press={scan}>
          扫描 Web 登录二维码
        </MobileButton>
      </View>
    </MobilePage>
  );
}
