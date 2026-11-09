import { Card, CardActionArea, Chip, LinearProgress } from '@mui/material';
import { Link, useLocation } from 'react-router-dom';
import { Icon } from '../components/common';
import { duration, resolution } from '../lib/format';
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

export function MediaCard({ item }: { item: Movie | Series }) {
  const location = useLocation();
  const movie = 'duration' in item ? item : null;
  const series = 'episode_count' in item ? item : null;
  const title = movie?.series_id ? `${movie.series_title} · ${movie.title}` : item.title;
  const status = item.metadata?.status;
  const badge = status === 'needs_review' ? '待确认匹配' : status === 'failed' ? '刮削失败' : '';
  const path = `/${series ? 'series' : 'movies'}/${item.id}`;
  return (
    <Card variant="outlined">
      <CardActionArea
        component={Link}
        to={path}
        state={{ kind: movie?.type || (series ? 'tv' : 'movie'), from: location.pathname + location.search }}
        className="movie-card"
        aria-label={`查看 ${title}`}
      >
        <span className="movie-artwork">
          <span className="movie-placeholder">
            <Icon name={series ? 'tv' : 'film'} />
          </span>
          <Poster url={item.poster_url} title={title} className="movie-poster" />
          {badge && (
            <Chip
              component="span"
              className="movie-metadata-badge"
              size="small"
              label={badge}
              color="warning"
            />
          )}
          {movie && (
            <Chip
              component="span"
              className="movie-quality"
              size="small"
              color="primary"
              label={resolution(movie)}
            />
          )}
          {movie?.watched && (
            <Chip component="span" className="movie-watched" size="small" label="已看" color="success" />
          )}
          {movie && movie.position > 0 && movie.duration > 0 && !movie.watched && (
            <LinearProgress
              variant="determinate"
              className="movie-watch-progress"
              value={Math.min(100, (100 * movie.position) / movie.duration)}
              aria-label="观看进度"
            />
          )}
        </span>
        <span className="movie-title">{title}</span>
        <span className="movie-meta">
          {item.year ? `${item.year} · ` : ''}
          {series
            ? `${series.season_count} 季 · ${series.episode_count} 个文件`
            : duration(movie?.duration || 0)}
        </span>
        {series && (
          <span className="movie-meta">
            已看 {series.watched_count} / {series.episode_count}
          </span>
        )}
      </CardActionArea>
    </Card>
  );
}
