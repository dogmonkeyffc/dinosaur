import { createContext, useContext, type ReactNode } from 'react';
import { useAuth } from '../auth/AuthProvider';
import { api, ApiError } from '../lib/api';
import { useResource } from '../lib/hooks';
import type { Library } from '../lib/types';

interface Libraries {
  libraries: Library[];
  error: string;
  loading: boolean;
  reload: () => void;
  scan: (items: Library[]) => Promise<void>;
}
const Context = createContext<Libraries>(null!);
export const useLibraries = () => useContext(Context);
export function LibraryProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth();
  const resource = useResource<{ libraries: Library[] }>(
    user?.role === 'admin' ? '/api/admin/libraries' : '/api/libraries',
    3000,
  );
  async function scan(items: Library[]) {
    if (!items.length) throw new Error('请先在媒体库设置中添加对应目录。');
    for (const item of items) {
      try {
        await api(`/api/admin/libraries/${item.id}/scan`, { method: 'POST' });
      } catch (error) {
        if (!(error instanceof ApiError) || error.code !== 'scan_running') throw error;
      }
    }
    resource.reload();
  }
  return (
    <Context.Provider value={{ libraries: resource.data?.libraries || [], ...resource, scan }}>
      {children}
    </Context.Provider>
  );
}
