import { resourceURL, serverURL } from './url';

const messages: Record<string, string> = {
  unauthorized: '登录已过期，请重新扫码登录。',
  pairing_expired: '配对码或二维码已过期、已使用，请重新生成。',
  invalid_pairing: '配对信息无效，请检查六位数字。',
  rate_limited: '操作过于频繁，请稍后重试。',
  forbidden: '当前账号没有此操作权限。',
};

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
  ) {
    super(messages[code] || `请求失败（${status}）。`);
  }
}

export class Transport {
  readonly origin: string;
  private abort = new AbortController();
  constructor(
    server: string,
    private token?: string,
    private fetcher: typeof fetch = fetch,
  ) {
    this.origin = serverURL(server);
  }
  dispose() {
    this.abort.abort();
  }
  headers(): Record<string, string> {
    return this.token ? { Authorization: `Bearer ${this.token}` } : {};
  }
  resource(path: string) {
    return { uri: resourceURL(this.origin, path), headers: this.headers() };
  }
  text(path: string, signal?: AbortSignal) {
    return this.request<string>(path, 'GET', undefined, signal, true);
  }
  async request<T>(
    path: string,
    method = 'GET',
    body?: unknown,
    signal?: AbortSignal,
    text = false,
  ): Promise<T> {
    if (this.abort.signal.aborted) throw new Error('连接已关闭。');
    const controller = new AbortController();
    const cancel = () => controller.abort();
    this.abort.signal.addEventListener('abort', cancel, { once: true });
    signal?.addEventListener('abort', cancel, { once: true });
    if (signal?.aborted) controller.abort();
    const timeout = setTimeout(cancel, 16000);
    try {
      const response = await this.fetcher(resourceURL(this.origin, path), {
        method,
        credentials: 'omit',
        redirect: 'error',
        signal: controller.signal,
        headers: { ...this.headers(), ...(body === undefined ? {} : { 'Content-Type': 'application/json' }) },
        body: body === undefined ? undefined : JSON.stringify(body),
      });
      if (controller.signal.aborted) throw new Error('请求已取消。');
      const data = response.ok && text ? await response.text() : await response.json();
      if (controller.signal.aborted) throw new Error('请求已取消。');
      if (!response.ok) throw new ApiError(response.status, data.error?.code || 'request_failed');
      return data as T;
    } finally {
      clearTimeout(timeout);
      this.abort.signal.removeEventListener('abort', cancel);
      signal?.removeEventListener('abort', cancel);
    }
  }
}
