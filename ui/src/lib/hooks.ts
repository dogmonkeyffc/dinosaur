import { useCallback, useEffect, useRef, useState } from 'react';
import { api } from './api';
import { message } from './errors';
import { flushProgress } from './progress';

export function useResource<T>(path: string | null, interval = 0) {
  const [data, setData] = useState<T>();
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(Boolean(path));
  const [revision, setRevision] = useState(0);
  const reload = useCallback(() => setRevision((value) => value + 1), []);
  const previousPath = useRef(path);
  useEffect(() => {
    const controller = new AbortController();
    if (previousPath.current !== path) setData(undefined);
    previousPath.current = path;
    if (!path) {
      setData(undefined);
      setLoading(false);
      setError('');
      return;
    }
    setLoading(true);
    setError('');
    flushProgress()
      .then(() => api<T>(path, { signal: controller.signal }))
      .then((result) => {
        if (!controller.signal.aborted) setData(result);
      })
      .catch((reason) => {
        if (!controller.signal.aborted && reason.name !== 'AbortError') setError(message(reason));
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });
    return () => controller.abort();
  }, [path, revision]);
  useEffect(() => {
    if (!path || !interval) return;
    const timer = window.setInterval(reload, interval);
    return () => clearInterval(timer);
  }, [path, interval, reload]);
  const current = previousPath.current === path;
  return {
    data: current ? data : undefined,
    error: current ? error : '',
    loading: current ? loading : Boolean(path),
    reload,
  };
}

export function useAction() {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const active = useRef(true);
  useEffect(() => {
    active.current = true;
    return () => {
      active.current = false;
    };
  }, []);
  async function run(action: () => Promise<void>) {
    if (busy) return;
    setBusy(true);
    setError('');
    try {
      await action();
    } catch (reason) {
      if (active.current && !(reason instanceof DOMException && reason.name === 'AbortError'))
        setError(message(reason));
    } finally {
      if (active.current) setBusy(false);
    }
  }
  return { busy, error, run, clearError: () => setError('') };
}

export function useDebounced<T>(value: T, delay = 250) {
  const [result, setResult] = useState(value);
  useEffect(() => {
    const timer = window.setTimeout(() => setResult(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);
  return result;
}
