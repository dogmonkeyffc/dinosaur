import { useEffect, useState } from 'react';
import type { Account, MediaLibrary } from '@dinosaur/contracts';
import { useClient, useMediaPage } from '@dinosaur/native-kit';
import { homeLibraries } from '../settings/homeVisibility';
import type { HomeTab } from './HomeHeader';

export function useMobileCatalog(account: Account, tab: HomeTab, hidden: number[], ready: boolean) {
  const client = useClient(account);
  const [libraries, setLibraries] = useState<MediaLibrary[]>([]);
  const [error, setError] = useState('');
  const [revision, setRevision] = useState(0);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [libraryID, setLibraryID] = useState<number>();
  const [page, setPage] = useState(1);
  const [query, setQuery] = useState('');
  const available = tab === 'home' ? (ready ? homeLibraries(libraries, hidden) : []) : libraries;
  const library = available.find((item) => item.id === libraryID) ?? available[0];
  const kind = library?.type ?? 'movie';
  const list = useMediaPage(client, kind, page, query, library?.id ?? null);
  useEffect(() => {
    let current = true;
    setError('');
    setLoading(true);
    client
      .libraries()
      .then(({ libraries: next }) => {
        if (current) setLibraries(next);
      })
      .catch((reason) => {
        if (current) setError(reason instanceof Error ? reason.message : '无法读取媒体库。');
      })
      .finally(() => {
        if (current) setLoading(false);
      });
    return () => {
      current = false;
    };
  }, [client, revision]);
  useEffect(() => {
    if (!loading && !list.loading) setRefreshing(false);
  }, [loading, list.loading]);
  function refresh() {
    if (refreshing) return;
    setRefreshing(true);
    setLoading(true);
    setRevision((value) => value + 1);
    list.reload();
  }
  function selectLibrary(id: number) {
    setLibraryID(id);
    setPage(1);
  }
  return {
    client,
    libraries,
    available,
    library,
    kind,
    list,
    page,
    setPage,
    query,
    setQuery,
    error,
    loading,
    refreshing,
    refresh,
    selectLibrary,
  };
}

export type MobileCatalog = ReturnType<typeof useMobileCatalog>;
