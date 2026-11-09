import { Button, LinearProgress, Paper } from '@mui/material';
import { useEffect } from 'react';
import { useAuth } from '../auth/AuthProvider';
import { ButtonLink } from '../components/ButtonLink';
import { ErrorMessage, Heading, Icon, Loading } from '../components/common';
import { clock } from '../lib/format';
import { useResource } from '../lib/hooks';
import type { Movie, Page, Series } from '../lib/types';
import { useLibraries } from '../libraries/LibraryProvider';
import { MediaCard, Poster } from './MediaCard';

function Shelf({ title, path, items }: { title: string; path: string; items: (Movie | Series)[] }) {
  if (!items.length) return null;
  return (
    <section className="home-shelf" aria-label={title}>
      <div className="shelf-heading">
        <h2>{title}</h2>
        <ButtonLink to={path}>查看全部{title}</ButtonLink>
      </div>
      <div className="movie-grid">
        {items.map((item) => (
          <MediaCard key={item.id} item={item} />
        ))}
      </div>
    </section>
  );
}

export function HomePage() {
  const { user } = useAuth();
  const libraries = useLibraries();
  const movies = useResource<Page<Movie>>('/api/media?page_size=6');
  const series = useResource<Page<Series>>('/api/series?page_size=6');
  const scanSignature = libraries.libraries
    .map((item) => `${item.id}:${item.media_count}:${item.scan?.status}`)
    .join(',');
  useEffect(() => {
    movies.reload();
    series.reload();
  }, [scanSignature, movies.reload, series.reload]);
  const featured = movies.data?.items.find((item) => item.position > 0 && !item.watched);
  const loading = movies.loading || series.loading || libraries.loading;
  const error = movies.error || series.error || libraries.error;
  const hasMedia = Boolean(movies.data?.total || series.data?.total);
  return (
    <section className="home-page">
      <Heading title="看片室" />
      <ErrorMessage error={error} />
      {error && (
        <Button
          onClick={() => {
            movies.reload();
            series.reload();
            libraries.reload();
          }}
        >
          重试读取
        </Button>
      )}
      {loading && !hasMedia && <Loading />}
      {featured && (
        <Paper component="section" className="resume-reel" aria-label="继续观看电影">
          <div className="resume-copy">
            <h2>{featured.title}</h2>
            <p>
              已观看 {clock(featured.position)}
              {featured.duration > 0 ? ` / ${clock(featured.duration)}` : ''}
            </p>
            {featured.duration > 0 && (
              <LinearProgress
                variant="determinate"
                value={Math.min(100, (100 * featured.position) / featured.duration)}
                aria-label="观看进度"
              />
            )}
            <ButtonLink to={`/watch/${featured.id}`} state={{ kind: 'movie', from: '/' }} variant="contained">
              继续播放
            </ButtonLink>
          </div>
          <Poster url={featured.poster_url} title={featured.title} className="resume-poster" />
        </Paper>
      )}
      <Shelf title="电影" path="/movies" items={movies.data?.items || []} />
      <Shelf title="剧集" path="/series" items={series.data?.items || []} />
      {!loading && !hasMedia && !error && (
        <div className="empty-state">
          <Icon name="film" />
          <h2>片架还是空的</h2>
          {user?.role === 'admin' && (
            <ButtonLink to="/settings" variant="contained">
              打开媒体库设置
            </ButtonLink>
          )}
        </div>
      )}
    </section>
  );
}
