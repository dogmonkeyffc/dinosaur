import { Link, useLocation } from 'react-router-dom';
import { Icon } from '../components/common';
import { clock, duration, resolution } from '../lib/format';
import type { Movie, Series } from '../lib/types';

export function Poster({ url, title, className }: { url?: string; title: string; className: string }) {
  return url ? (
    <img
      key={url}
      className={className}
      src={url}
      alt={`${title} 海报`}
      loading="lazy"
      onError={(event) => {
        event.currentTarget.hidden = true;
      }}
    />
  ) : null;
}

export function MediaCard({ item, resume = false }: { item: Movie | Series; resume?: boolean }) {
  const location = useLocation();
  const movie = 'duration' in item ? item : null;
  const series = 'episode_count' in item ? item : null;
  const title = movie?.series_id ? `${movie.series_title} · ${movie.title}` : item.title;
  const status = item.metadata?.status;
  const badge = status === 'needs_review' ? '待确认匹配' : status === 'failed' ? '刮削失败' : '';
  const path = resume ? `/watch/${item.id}` : `/${series ? 'series' : 'movies'}/${item.id}`;
  return (
    <Link
      to={path}
      state={{ kind: movie?.type || (series ? 'tv' : 'movie'), from: location.pathname + location.search }}
      className="movie-card"
      aria-label={`${resume ? '继续观看' : '查看'} ${title}`}
    >
      <span className="movie-artwork">
        <span className="movie-placeholder">
          <Icon name={series ? 'tv' : 'film'} />
          <span>DINOSAUR COLLECTION</span>
        </span>
        <Poster url={item.poster_url} title={title} className="movie-poster" />
        {badge && <span className="movie-metadata-badge">{badge}</span>}
        {movie && <span className="movie-quality">{resolution(movie)}</span>}
        {movie?.watched && <span className="movie-watched">已看</span>}
        <span className="movie-artwork-caption">{item.library_name}</span>
        {movie && movie.position > 0 && movie.duration > 0 && !movie.watched && (
          <progress
            className="movie-watch-progress"
            value={movie.position}
            max={movie.duration}
            aria-label="观看进度"
          />
        )}
      </span>
      <span className="movie-title">{title}</span>
      <span className="movie-meta">
        {item.year || '年份未知'} ·{' '}
        {series
          ? `${series.season_count} 季 · ${series.episode_count} 个文件`
          : resume
            ? `已看 ${clock(movie?.position)}`
            : duration(movie?.duration || 0)}
      </span>
      {series && (
        <span className="movie-meta">
          已看 {series.watched_count} / {series.episode_count}
        </span>
      )}
    </Link>
  );
}
