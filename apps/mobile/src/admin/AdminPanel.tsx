import { useEffect, useRef, useState } from 'react';
import { BackHandler, Platform, Pressable, StatusBar, Text, View } from 'react-native';
import { WebView } from 'react-native-webview';
import type { Account } from '@dinosaur/contracts';
import type { Client } from '@dinosaur/client';
import { Busy, ErrorText, PlaybackIcon } from '@dinosaur/native-kit';
import { pageStyles as css } from '../ui/pageStyles';
import { adminBootstrap, adminNavigationAllowed } from './navigation';

export function AdminPanel({
  account,
  client,
  close,
}: {
  account: Account;
  client: Client;
  close: () => void;
}) {
  const browser = useRef<WebView>(null);
  const canGoBack = useRef(false);
  const bootstrapped = useRef(false);
  const [code, setCode] = useState('');
  const [error, setError] = useState('');
  const [revision, setRevision] = useState(0);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    let active = true;
    setCode('');
    bootstrapped.current = false;
    setError('');
    setLoading(true);
    if (account.user.role !== 'admin') {
      setError('当前账号没有管理权限。');
      setLoading(false);
      return;
    }
    client
      .adminWebLogin()
      .then((result) => {
        adminBootstrap(client.origin, result.code);
        if (active) setCode(result.code);
      })
      .catch((reason) => {
        if (active) {
          setError(reason instanceof Error ? reason.message : '无法打开管理');
          setLoading(false);
        }
      });
    return () => {
      active = false;
    };
  }, [client, account.user.role, revision]);
  useEffect(
    () => () => {
      void client.closeAdminWeb().catch(() => {});
    },
    [client],
  );
  function back() {
    if (canGoBack.current) browser.current?.goBack();
    else close();
  }
  useEffect(() => {
    const listener = BackHandler.addEventListener('hardwareBackPress', () => {
      back();
      return true;
    });
    return () => listener.remove();
  });
  return (
    <View style={[css.screen, { paddingTop: Platform.OS === 'android' ? StatusBar.currentHeight : 52 }]}>
      <StatusBar barStyle="light-content" />
      <View style={css.toolbar}>
        <Pressable accessibilityRole="button" accessibilityLabel="返回" onPress={back} style={css.back}>
          <PlaybackIcon name="back" />
        </Pressable>
        <Text accessibilityRole="header" style={css.title}>
          管理
        </Text>
        <Pressable accessibilityRole="button" accessibilityLabel="关闭管理" onPress={close} style={css.back}>
          <PlaybackIcon name="close" />
        </Pressable>
      </View>
      {loading && <Busy />}
      <ErrorText error={error} />
      {!!error && (
        <Pressable
          accessibilityRole="button"
          onPress={() => setRevision((value) => value + 1)}
          style={css.button}
        >
          <Text style={css.buttonText}>重试</Text>
        </Pressable>
      )}
      {!!code && !error && (
        <WebView
          ref={browser}
          key={`${account.key}:${revision}`}
          style={{ flex: 1, backgroundColor: '#111113' }}
          source={adminBootstrap(client.origin, code)}
          incognito
          cacheEnabled={false}
          sharedCookiesEnabled={false}
          thirdPartyCookiesEnabled={false}
          mixedContentMode="never"
          allowFileAccess={false}
          javaScriptCanOpenWindowsAutomatically={false}
          setSupportMultipleWindows={false}
          // Let our same-origin check reject links without WebView opening external apps.
          originWhitelist={['*']}
          onShouldStartLoadWithRequest={(request) => adminNavigationAllowed(client.origin, request.url)}
          onNavigationStateChange={(state) => {
            canGoBack.current = state.canGoBack;
          }}
          onLoadEnd={(event) => {
            if (!bootstrapped.current && event.nativeEvent.url === client.origin + '/settings') {
              bootstrapped.current = true;
              browser.current?.clearHistory?.();
              canGoBack.current = false;
            }
            setLoading(false);
          }}
          onError={() => {
            setError('无法加载管理页面，请检查服务器连接。');
            setLoading(false);
          }}
          onHttpError={(event) => {
            setError(`管理页面请求失败（${event.nativeEvent.statusCode}）`);
            setLoading(false);
          }}
        />
      )}
    </View>
  );
}
