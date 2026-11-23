export * from './history';

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

export interface MediaLibrary {
  id: number;
  name: string;
  type: MediaKind;
}
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
export interface Work {
  id: number;
  title: string;
  library_id: number;
  poster_url?: string;
  episode_count: number;
  season_count: number;
  watched_count: number;
  member_count: number;
}
export interface WorkDetail {
  work: Work;
  members: Series[];
}
export interface Series {
  work_id?: number;
  member_count?: number;
  library_id?: number;
  season_count?: number;
  watched_count?: number;
  id: number;
  title: string;
  year: number;
  poster_url?: string;
  metadata?: { overview: string };
  episode_count: number;
}
export interface SeriesDetail {
  work?: Work;
  series: Series;
  seasons: { number: number; episodes: Movie[] }[];
  resume?: Movie;
}
export interface PlaybackInfo {
  selected_subtitle?: number | null;
  item: Movie;
  next_episode?: Movie;
  plan: { mode: 'direct' | 'remux' | 'transcode'; audio: number; start: number; subtitle?: number };
  stream_url: string;
  resume_position: number;
  audio_tracks: Track[];
  subtitles: {
    id: number;
    label: string;
    kind?: 'embedded' | 'external';
    delivery?: 'text' | 'burn';
    language?: string;
    codec?: string;
  }[];
}
