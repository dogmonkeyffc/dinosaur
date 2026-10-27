export interface User {
  id: number;
  username: string;
  role: 'admin' | 'user';
}
export interface ServerInfo {
  id: string;
  name: string;
  api_version: number;
  capabilities: string[];
}
export interface ConnectedServer extends ServerInfo {
  url: string;
}
export interface DeviceSession {
  token: string;
  user: User;
  expires_at: number;
}
export interface Account extends DeviceSession {
  key: string;
  server: ConnectedServer;
}
export interface Device {
  id: string;
  name: string;
  created_at: number;
  expires_at: number;
  current: boolean;
}
export interface Pairing {
  pairing_id: string;
  name: string;
  expires_at: number;
}
export interface TVPairing extends Pairing {
  code: string;
  device_code: string;
  interval: number;
}
export type PollResult = DeviceSession | { status: 'pending'; interval: number };
export interface Page<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}
export type MediaKind = 'movie' | 'tv';
export interface Track {
  index: number;
  type: string;
  codec: string;
  language: string;
  title: string;
}
export interface Movie {
  id: number;
  title: string;
  type: MediaKind;
  duration: number;
  position: number;
  watched: boolean;
  year: number;
  poster_url?: string;
  metadata?: { overview: string };
  series_id?: number;
  series_title?: string;
  season_number?: number;
}
export interface Series {
  id: number;
  title: string;
  year: number;
  poster_url?: string;
  metadata?: { overview: string };
  episode_count: number;
}
export interface SeriesDetail {
  series: Series;
  seasons: { number: number; episodes: Movie[] }[];
  resume?: Movie;
}
export interface PlaybackInfo {
  item: Movie;
  next_episode?: Movie;
  plan: { mode: 'direct' | 'remux' | 'transcode'; audio: number; start: number };
  stream_url: string;
  resume_position: number;
  audio_tracks: Track[];
  subtitles: { id: number; label: string }[];
}
