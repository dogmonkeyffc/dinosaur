import { useState } from 'react';
import { Modal, Platform, Pressable, ScrollView, StatusBar, Switch, Text, View } from 'react-native';
import type { Account, MediaLibrary } from '@dinosaur/contracts';
import { Busy, ErrorText } from '@dinosaur/native-kit';
import type { useHomeVisibility } from './useHomeVisibility';
import { settingsStyles as css } from './settingsStyles';
import { palette } from '../catalog/homeStyles';
import appConfig from '../../app.json';
import { dp, uiScale } from '../ui/density';
import type { Client } from '@dinosaur/client';
import { HistoryRecordingSettings } from './HistoryRecordingSettings';

export function AccountSettings({
  account,
  client,
  libraries,
  visibility,
  close,
  accounts,
  pair,
  loading,
  error,
  retry,
}: {
  account: Account;
  client: Client;
  libraries: MediaLibrary[];
  visibility: ReturnType<typeof useHomeVisibility>;
  close: () => void;
  accounts: () => void;
  pair: () => void;
  loading: boolean;
  error: string;
  retry: () => void;
}) {
  const [page, setPage] = useState<'account' | 'libraries' | 'history'>('account');
  function back() {
    if (page !== 'account') setPage('account');
    else close();
  }
  function choose(action: () => void) {
    close();
    action();
  }
  return (
    <Modal transparent animationType="slide" statusBarTranslucent onRequestClose={back}>
      <View
        style={[
          css.backdrop,
          { paddingTop: (Platform.OS === 'android' ? (StatusBar.currentHeight ?? 24) : 52) + dp(10) },
        ]}
      >
        <View accessibilityViewIsModal style={css.panel}>
          <View style={css.toolbar}>
            {page !== 'account' ? (
              <Pressable
                accessibilityRole="button"
                accessibilityLabel="返回账号设置"
                hitSlop={4}
                style={css.iconButton}
                onPress={back}
              >
                <Text style={css.icon}>←</Text>
              </Pressable>
            ) : (
              <View style={css.iconButton} />
            )}
            <Text numberOfLines={1} style={css.server}>
              {account.server.name}
            </Text>
            <Pressable
              accessibilityRole="button"
              accessibilityLabel="关闭设置"
              hitSlop={4}
              style={css.iconButton}
              onPress={close}
            >
              <Text style={css.icon}>×</Text>
            </Pressable>
          </View>
          <ScrollView contentContainerStyle={css.content}>
            {page === 'account' ? (
              <>
                <View style={css.profile}>
                  <View style={css.avatarRing}>
                    <View style={css.avatar}>
                      <Text style={css.initial}>{account.user.username.slice(0, 1).toUpperCase()}</Text>
                    </View>
                  </View>
                  <Text accessibilityRole="header" style={css.username}>
                    {account.user.username}
                  </Text>
                </View>
                <View style={css.group}>
                  <SettingsRow icon="⇄" title="切换账号" press={() => choose(accounts)} />
                </View>
                <View style={css.group}>
                  <SettingsRow icon="▣" title="登录电视" press={() => choose(pair)} />
                  <SettingsRow icon="⚙" title="首页显示" border press={() => setPage('libraries')} />
                  <SettingsRow icon="◷" title="记录观看" border press={() => setPage('history')} />
                </View>
                <Text style={css.version}>Dinosaur {appConfig.expo.version}</Text>
              </>
            ) : page === 'history' ? (
              <HistoryRecordingSettings client={client} />
            ) : (
              <>
                <Text accessibilityRole="header" style={css.title}>
                  首页显示
                </Text>
                <ErrorText error={visibility.error || error} />
                {(loading || (!visibility.ready && !visibility.error)) && <Busy />}
                {((!visibility.ready && !!visibility.error) || !!error) && (
                  <Pressable
                    accessibilityRole="button"
                    style={css.retry}
                    onPress={() => {
                      visibility.retry();
                      retry();
                    }}
                  >
                    <Text style={css.label}>重试</Text>
                  </Pressable>
                )}
                {!loading && !error && libraries.length === 0 && <Text style={css.label}>暂无媒体库</Text>}
                <View style={css.group}>
                  {libraries.map((library, index) => (
                    <View key={library.id} style={[css.row, index > 0 && css.rowBorder]}>
                      <Text style={css.label}>{library.name}</Text>
                      <Switch
                        style={{ transform: [{ scale: uiScale }] }}
                        accessibilityLabel={`首页显示${library.name}`}
                        disabled={!visibility.ready || visibility.busy}
                        value={!visibility.hidden.includes(library.id)}
                        onValueChange={() => visibility.toggle(library.id)}
                        trackColor={{ false: '#44464e', true: '#4c638b' }}
                        thumbColor={palette.accent}
                      />
                    </View>
                  ))}
                </View>
              </>
            )}
          </ScrollView>
        </View>
      </View>
    </Modal>
  );
}

function SettingsRow({
  icon,
  title,
  press,
  border = false,
}: {
  icon: string;
  title: string;
  press: () => void;
  border?: boolean;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      onPress={press}
      style={({ pressed }) => [css.row, border && css.rowBorder, pressed && { opacity: 0.65 }]}
    >
      <Text style={css.rowIcon}>{icon}</Text>
      <Text style={css.label}>{title}</Text>
      <Text style={css.chevron}>›</Text>
    </Pressable>
  );
}
