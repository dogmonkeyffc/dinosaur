import type {
  Device,
  DeviceSession,
  MediaLibrary,
  Movie,
  Page,
  Pairing,
  PlaybackInfo,
  PollResult,
  ServerInfo,
  Series,
  SeriesDetail,
  TVPairing,
  User,
  WorkDetail,
  WatchHistoryEntry,
} from '@dinosaur/contracts';
import { Transport } from './transport';
export { ApiError, Transport } from './transport';
export { loginQR, resourceURL, serverURL } from './url';

export class Client extends Transport {
  info(signal?: AbortSignal) {
    return this.request<ServerInfo>('/api/server', 'GET', undefined, signal);
  }
  me() {
    return this.request<{ user: User }>('/api/auth/me');
  }
  exchangeLoginCode(code: string, name: string) {
    return this.request<DeviceSession>('/api/devices/login-code/exchange', 'POST', { code, name });
  }
  exchange(ticket: string, name: string) {
    return this.request<DeviceSession>('/api/devices/qr/exchange', 'POST', { ticket, name });
  }
  createPairing(name: string) {
    return this.request<TVPairing>('/api/devices/pairings', 'POST', { name });
  }
  poll(device_code: string) {
    return this.request<PollResult>('/api/devices/pairings/poll', 'POST', { device_code });
  }
  cancelPairing(device_code: string) {
    return this.request('/api/devices/pairings/cancel', 'POST', { device_code });
  }
  lookup(code: string) {
    return this.request<Pairing>('/api/devices/pairings/lookup', 'POST', { code });
  }
  approve(pairing_id: string) {
    return this.request<{ ok: boolean; user: User }>('/api/devices/pairings/approve', 'POST', { pairing_id });
  }
  devices() {
    return this.request<{ devices: Device[] }>('/api/devices');
  }
  revoke(id: string) {
    return this.request('/api/devices/' + encodeURIComponent(id), 'DELETE');
  }
  logout() {
    return this.request('/api/auth/logout', 'POST');
  }
  libraries() {
    return this.request<{ libraries: MediaLibrary[] }>('/api/libraries');
  }
  movies(page = 1, q = '', libraryID?: number) {
    return this.request<Page<Movie>>(`/api/media?${catalogQuery(page, q, libraryID)}`);
  }
  series(page = 1, q = '', libraryID?: number) {
    return this.request<Page<Series>>(`/api/series?${catalogQuery(page, q, libraryID)}`);
  }
  movie(id: number) {
    return this.request<{ item: Movie }>(`/api/media/${id}`);
  }
  detail(id: number) {
    return this.request<SeriesDetail>(`/api/series/${id}`);
  }
  work(id: number) {
    return this.request<WorkDetail>(`/api/works/${id}`);
  }
  recent() {
    return this.request<{ items: Movie[] }>('/api/playback/continue');
  }
  history(page = 1) {
    return this.request<Page<WatchHistoryEntry>>(`/api/history?page=${page}&page_size=24`);
  }
  playback(id: number, mode = 'original', start = 0, audio?: number, subtitle?: number | 'auto') {
    const query = new URLSearchParams({ mode, start: String(Math.max(0, start)) });
    if (audio !== undefined) query.set('audio', String(audio));
    if (subtitle !== undefined) query.set('subtitle', String(subtitle));
    return this.request<PlaybackInfo>(`/api/playback/${id}?${query}`);
  }
  progress(id: number, position: number, watched?: boolean) {
    return this.request(`/api/playback/${id}/progress`, 'PUT', {
      position,
      ...(watched === undefined ? {} : { watched }),
    });
  }
}

function catalogQuery(page: number, q: string, libraryID?: number) {
  const query = `page=${page}&page_size=24&q=${encodeURIComponent(q)}`;
  return libraryID === undefined ? query : `${query}&library_id=${libraryID}`;
}
