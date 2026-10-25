import { errorMessages } from './errors';
import type { Session } from './types';

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
  ) {
    super(errorMessages[code] || '请求失败，请稍后重试。');
  }
}

let session: Session | null = null;
let epoch = 0;
let sessionAbort = new AbortController();

export function setSession(value: Session | null) {
  sessionAbort.abort();
  sessionAbort = new AbortController();
  session = value;
  epoch++;
}

export const currentSession = () => ({ session, epoch });

interface Options {
  method?: string;
  body?: unknown;
  signal?: AbortSignal;
  headers?: Record<string, string>;
  keepalive?: boolean;
}

export async function api<T = Record<string, never>>(path: string, options: Options = {}): Promise<T> {
  const { method = 'GET', body, signal, keepalive = false } = options;
  const revision = epoch;
  const headers = { ...options.headers };
  if (body !== undefined) headers['Content-Type'] = 'application/json';
  if (session && !['GET', 'HEAD'].includes(method)) headers['X-CSRF-Token'] = session.csrf_token;
  const signals = [AbortSignal.timeout(16000), sessionAbort.signal];
  if (signal) signals.push(signal);
  const combined = AbortSignal.any(signals);
  try {
    const response = await fetch(path, {
      method,
      headers,
      credentials: 'same-origin',
      signal: combined,
      keepalive,
      body: body === undefined ? undefined : JSON.stringify(body),
    });
    const data = await response.json();
    if (!response.ok) {
      if (response.status === 401 && session && revision === epoch)
        window.dispatchEvent(new Event('session-expired'));
      throw new ApiError(response.status, data.error?.code || 'internal_error');
    }
    if (revision !== epoch) throw new DOMException('Session changed', 'AbortError');
    return data as T;
  } catch (error) {
    if (combined.aborted) {
      if (combined.reason?.name === 'TimeoutError') throw new Error('连接超时，请检查服务器状态后重试。');
      throw new DOMException('Request cancelled', 'AbortError');
    }
    if (error instanceof TypeError) throw new Error('无法连接服务器，请检查网络后重试。');
    throw error;
  }
}
