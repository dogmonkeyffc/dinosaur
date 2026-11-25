import { useState } from 'react';
import { Text, TextInput, View } from 'react-native';
import type { Account, Pairing } from '@dinosaur/contracts';
import { ErrorText, useClient, useTask } from '@dinosaur/native-kit';
import { MobileButton, MobilePage } from '../ui/MobilePage';
import { pageStyles as css } from '../ui/pageStyles';
import { palette } from '../catalog/homeStyles';

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
    <MobilePage title="登录电视" close={close}>
      <View style={css.card}>
        <View style={css.account}>
          <View style={css.avatar}>
            <Text style={css.initial}>{account.user.username.slice(0, 1).toUpperCase()}</Text>
          </View>
          <View style={css.identity}>
            <Text style={css.name}>{account.user.username}</Text>
            <Text style={css.server} numberOfLines={1}>
              {account.server.name || account.server.url}
            </Text>
          </View>
        </View>
        <MobileButton disabled={task.busy} press={switchAccount}>
          切换账号
        </MobileButton>
      </View>
      <View style={css.card}>
        <Text style={css.device}>{done ? '✓' : '▣'}</Text>
        {done ? (
          <>
            <Text style={css.message}>已授权 {pairing?.name}</Text>
            <MobileButton primary press={close}>
              完成
            </MobileButton>
          </>
        ) : pairing ? (
          <>
            <Text style={css.message}>{pairing.name}</Text>
            <MobileButton
              primary
              disabled={task.busy}
              press={() =>
                task.run(async () => {
                  await client.approve(pairing.pairing_id);
                  setDone(true);
                })
              }
            >
              确认授权
            </MobileButton>
            <MobileButton
              disabled={task.busy}
              press={() => {
                setPairing(undefined);
                task.clear();
              }}
            >
              重新输入
            </MobileButton>
          </>
        ) : (
          <>
            <TextInput
              accessibilityLabel="电视六位配对码"
              placeholder="000000"
              placeholderTextColor={palette.muted}
              keyboardType="number-pad"
              maxLength={6}
              value={code}
              editable={!task.busy}
              style={css.input}
              onChangeText={(value) => setCode(value.replace(/\D/g, ''))}
            />
            <MobileButton
              primary
              disabled={task.busy || code.length !== 6}
              press={() =>
                task.run(async () => {
                  setPairing(await client.lookup(code));
                })
              }
            >
              查找电视
            </MobileButton>
          </>
        )}
        <ErrorText error={task.error} />
      </View>
    </MobilePage>
  );
}
