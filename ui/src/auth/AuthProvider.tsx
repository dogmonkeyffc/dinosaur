import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from 'react';
import { api, setSession } from '../lib/api';
import { message } from '../lib/errors';
import { flushProgress } from '../lib/progress';
import type { Session, User } from '../lib/types';

type Status = 'loading' | 'error' | 'setup' | 'guest' | 'ready';
interface Auth {
  user: User | null;
  status: Status;
  error: string;
  boot: () => void;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  initialized: () => void;
  expired: () => void;
}
const Context = createContext<Auth>(null!);
export const useAuth = () => useContext(Context);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [status, setStatus] = useState<Status>('loading');
  const [error, setError] = useState('');
  const accept = (session: Session) => {
    setSession(session);
    setUser(session.user);
    setStatus('ready');
  };
  const expired = useCallback(() => {
    setSession(null);
    setUser(null);
    setStatus('guest');
  }, []);
  const boot = useCallback(async () => {
    setStatus('loading');
    setError('');
    try {
      const setup = await api<{ initialized: boolean }>('/api/setup');
      if (!setup.initialized) {
        setStatus('setup');
        return;
      }
      try {
        accept(await api<Session>('/api/auth/me'));
      } catch (reason) {
        if (reason instanceof Error && 'status' in reason && reason.status === 401) setStatus('guest');
        else throw reason;
      }
    } catch (reason) {
      setError(message(reason));
      setStatus('error');
    }
  }, []);
  useEffect(() => {
    boot();
    window.addEventListener('session-expired', expired);
    return () => window.removeEventListener('session-expired', expired);
  }, [boot, expired]);
  async function login(username: string, password: string) {
    accept(await api<Session>('/api/auth/login', { method: 'POST', body: { username, password } }));
  }
  async function logout() {
    await flushProgress();
    await api('/api/auth/logout', { method: 'POST' });
    expired();
  }
  return (
    <Context.Provider
      value={{ user, status, error, boot, login, logout, expired, initialized: () => setStatus('guest') }}
    >
      {children}
    </Context.Provider>
  );
}
