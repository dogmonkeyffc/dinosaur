import { useEffect, useRef, useState } from 'react';
import { Pressable, Switch, Text, View } from 'react-native';
import type { Client } from '@dinosaur/client';
import type { HistoryLibrarySetting } from '@dinosaur/contracts';
import { Busy, ErrorText } from '@dinosaur/native-kit';
import { settingsStyles as css } from './settingsStyles';
import { palette } from '../catalog/homeStyles';
import { uiScale } from '../ui/density';

export function HistoryRecordingSettings({ client }: { client: Client }) {
  const [items, setItems] = useState<HistoryLibrarySetting[]>();
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [revision, setRevision] = useState(0);
  const pending = useRef(false);
  const active = useRef(true);
  useEffect(() => {
    active.current = true;
    let cancelled = false;
    setLoading(true);
    setError('');
    client
      .historySettings()
      .then((result) => {
        if (!cancelled) setItems(result.libraries);
      })
      .catch((reason) => {
        if (!cancelled) setError(reason instanceof Error ? reason.message : '读取失败');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
      active.current = false;
    };
  }, [client, revision]);
  async function toggle(item: HistoryLibrarySetting) {
    if (pending.current) return;
    pending.current = true;
    setBusy(true);
    setError('');
    try {
      await client.setHistoryRecording(item.id, !item.record_history);
      if (active.current)
        setItems((previous) =>
          previous?.map((value) =>
            value.id === item.id ? { ...value, record_history: !item.record_history } : value,
          ),
        );
    } catch (reason) {
      if (active.current) setError(reason instanceof Error ? reason.message : '保存失败');
    } finally {
      pending.current = false;
      if (active.current) setBusy(false);
    }
  }
  return (
    <>
      <Text accessibilityRole="header" style={css.title}>
        记录观看
      </Text>
      <ErrorText error={error} />
      {loading && <Busy />}
      {!!error && !items && (
        <Pressable
          accessibilityRole="button"
          style={css.retry}
          onPress={() => setRevision((value) => value + 1)}
        >
          <Text style={css.label}>重试</Text>
        </Pressable>
      )}
      {!loading && !error && items?.length === 0 && <Text style={css.label}>暂无媒体库</Text>}
      <View style={css.group}>
        {items?.map((item, index) => (
          <View key={item.id} style={[css.row, index > 0 && css.rowBorder]}>
            <Text style={css.label}>{item.name}</Text>
            <Switch
              accessibilityLabel={`记录${item.name}`}
              value={item.record_history}
              disabled={busy || loading}
              onValueChange={() => toggle(item)}
              style={{ transform: [{ scale: uiScale }] }}
              trackColor={{ false: '#44464e', true: '#4c638b' }}
              thumbColor={palette.accent}
            />
          </View>
        ))}
      </View>
    </>
  );
}
