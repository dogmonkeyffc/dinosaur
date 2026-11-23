export interface WatchHistoryEntry {
  id: number;
  media_id?: number;
  title: string;
  original_title: string;
  year: number;
  watched_on: string;
  rating: number;
  review: string;
  source: 'manual' | 'douban' | 'local';
  source_id: string;
  poster_url?: string;
  source_url?: string;
}
