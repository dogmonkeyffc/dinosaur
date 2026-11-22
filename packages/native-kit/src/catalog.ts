import { useEffect, useState } from 'react';
import type { Movie, Series, SeriesDetail, MediaKind } from '@dinosaur/contracts';
import type { Client } from '@dinosaur/client';

export function useMediaPage(client: Client, kind: MediaKind, page: number, query: string, libraryID?: number | null) {
  const scope = `${kind}:${libraryID}`;
  const [loadedScope, setLoadedScope] = useState<string>();
  const [items, setItems] = useState<(Movie | Series)[]>([]);
  const [recent, setRecent] = useState<Movie[]>([]);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [revision, setRevision] = useState(0);
  useEffect(() => {
    if (libraryID === null) return;
    let active = true;
    setLoading(true);
    setError('');
    const timer = setTimeout(
      async () => {
        try {
          const [list, progress] = await Promise.all([
            kind === 'tv' ? client.series(page, query, libraryID) : client.movies(page, query, libraryID),
            client.recent(),
          ]);
          if (active) {
            setItems(list.items);
            setLoadedScope(scope);
            setTotal(list.total);
            setRecent(progress.items);
          }
        } catch (reason) {
          if (active) setError(reason instanceof Error ? reason.message : '无法读取媒体。');
        } finally {
          if (active) setLoading(false);
        }
      },
      query ? 300 : 0,
    );
    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, [client, kind, page, query, revision, libraryID, scope]);
  return {
    items: loadedScope === scope ? items : [],
    recent,
    total: loadedScope === scope ? total : 0,
    error: libraryID === null ? '' : error,
    loading: libraryID !== null && loading,
    reload: () => setRevision((value) => value + 1),
  };
}

export function useMediaDetail(client: Client, kind: MediaKind, id: number, enabled = true) {
  const [movie, setMovie] = useState<Movie>();
  const [series, setSeries] = useState<SeriesDetail>();
  const [error, setError] = useState('');
  useEffect(() => {
    if (!enabled) return;
    let active = true;
    setMovie(undefined);
    setSeries(undefined);
    setError('');
    const request = kind === 'tv' ? client.detail(id) : client.movie(id);
    request
      .then((value) => {
        if (!active) return;
        if ('series' in value) setSeries(value);
        else setMovie(value.item);
      })
      .catch((reason) => {
        if (active) setError(reason instanceof Error ? reason.message : '无法读取详情。');
      });
    return () => {
      active = false;
    };
  }, [client, kind, id, enabled]);
  return { movie, series, error };
}
