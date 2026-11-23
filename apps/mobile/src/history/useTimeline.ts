import { useEffect, useRef, useState } from 'react';
import type { Client } from '@dinosaur/client';
import type { WatchHistoryEntry } from '@dinosaur/contracts';
import { mergeTimeline } from './timeline';

export function useTimeline(client: Client) {
  const [items, setItems] = useState<WatchHistoryEntry[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(0);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState('');
  const active = useRef(true);
  const pending = useRef(false);
  const generation = useRef(0);
  const attempted = useRef(1);
  async function load(next: number, reset = false) {
    if (pending.current && !reset) return;
    if (reset) generation.current++;
    const current = generation.current;
    pending.current = true;
    attempted.current = next;
    setLoading(true);
    setError('');
    try {
      const result = await client.history(next);
      if (!active.current || current !== generation.current) return;
      setItems((previous) => (reset ? result.items : mergeTimeline(previous, result.items)));
      setPage(next);
      setTotal(result.total);
    } catch (reason) {
      if (active.current && current === generation.current) {
        setError(reason instanceof Error ? reason.message : '无法读取观看记录。');
      }
    } finally {
      if (active.current && current === generation.current) {
        pending.current = false;
        setLoading(false);
        setRefreshing(false);
      }
    }
  }
  useEffect(() => {
    active.current = true;
    load(1, true);
    return () => {
      active.current = false;
      generation.current++;
    };
  }, [client]);
  function refresh() {
    setRefreshing(true);
    load(1, true);
  }
  return {
    items,
    total,
    loading,
    refreshing,
    error,
    refresh,
    more: () => {
      if (!error && page > 0 && page * 24 < total) load(page + 1);
    },
    retry: () => load(attempted.current, attempted.current === 1),
  };
}
