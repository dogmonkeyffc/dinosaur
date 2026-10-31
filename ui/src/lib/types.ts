export type MediaKind = 'movie' | 'tv';
export interface User {
  id: number;
  username: string;
  role: 'admin' | 'user';
}
export interface Session {
  user: User;
  csrf_token: string;
}
export interface Scan {
  status: string;
  processed: number;
  discovered: number;
  errors: number;
}
export interface Library {
  id: number;
  name: string;
  type: MediaKind;
  root_path?: string;
  available?: boolean;
  media_count: number;
  scan?: Scan;
}
export interface Candidate {
  id: number;
  title: string;
  original_title: string;
  overview: string;
  year: number;
}
export interface Metadata {
  tmdb_id: number;
  status: string;
  overview: string;
  original_title: string;
  message: string;
  candidates?: Candidate[];
}
export interface Track {
  index: number;
  type: string;
  codec: string;
  language: string;
  title: string;
  channels?: number;
}
export interface Movie {
  id: number;
  type: MediaKind;
  title: string;
  year: number;
  filename: string;
  size: number;
  library_id: number;
  library_name: string;
  duration: number;
  format: string;
  video_codec: string;
  width: number;
  height: number;
  tracks: Track[];
  position: number;
  watched: boolean;
  poster_url?: string;
  probe_error: boolean;
  metadata?: Metadata;
  series_id?: number;
  series_title?: string;
  season_number?: number;
}
export interface Series {
  id: number;
  title: string;
  year: number;
  library_id: number;
  library_name: string;
  season_count: number;
  episode_count: number;
  watched_count: number;
  metadata?: Metadata;
  poster_url?: string;
}
export interface SeriesDetail {
  series: Series;
  folders: string[];
  seasons: { number: number; episodes: Movie[] }[];
  resume?: Movie;
}
export interface Page<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}
export interface Job {
  id: number;
  media_type: MediaKind;
  library_id: number;
  media_id: number;
  status: string;
  total: number;
  processed: number;
  matched: number;
  review: number;
  errors: number;
  skipped: number;
  message: string;
  created_at: number;
  ended_at: number;
}
export interface ScrapeRecord {
  media_id: number;
  series_id?: number;
  library_name: string;
  title: string;
  filename: string;
  tmdb_id: number;
  language?: string;
  status: string;
  message: string;
  updated_at: number;
  candidates?: Candidate[];
}
export interface PlaybackInfo {
  selected_subtitle?: number | null;
  item: Movie;
  next_episode?: Movie;
  plan: { mode: 'direct' | 'remux' | 'transcode'; audio: number; subtitle?: number };
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
