import { useEffect, useRef, useState } from 'react';
import * as SecureStore from 'expo-secure-store';
import { parseHiddenLibraries } from './homeVisibility';

export function useHomeVisibility(accountKey: string) {
  const key = `dinosaur.home-libraries.${accountKey}`;
  const [hidden, setHidden] = useState<number[]>([]);
  const [ready, setReady] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [revision, setRevision] = useState(0);
  const pending = useRef(false);
  const active = useRef(true);
  useEffect(() => {
    active.current = true;
    return () => {
      active.current = false;
    };
  }, []);
  useEffect(() => {
    let current = true;
    setReady(false);
    setError('');
    SecureStore.getItemAsync(key)
      .then((raw) => {
        if (!current) return;
        setHidden(parseHiddenLibraries(raw));
        setReady(true);
      })
      .catch(() => {
        if (current) setError('无法读取首页显示设置，请重试。');
      });
    return () => {
      current = false;
    };
  }, [key, revision]);
  async function toggle(id: number) {
    if (!ready || pending.current) return;
    pending.current = true;
    setBusy(true);
    setError('');
    const next = hidden.includes(id) ? hidden.filter((value) => value !== id) : [...hidden, id];
    try {
      await SecureStore.setItemAsync(key, JSON.stringify(next));
      if (active.current) setHidden(next);
    } catch {
      if (active.current) setError('保存失败，请重试。');
    } finally {
      pending.current = false;
      if (active.current) setBusy(false);
    }
  }
  return { hidden, ready, busy, error, toggle, retry: () => setRevision((value) => value + 1) };
}
