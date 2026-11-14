import { Button, Paper } from '@mui/material';
import { ButtonLink } from '../components/ButtonLink';
import { SeriesSeasons } from './SeriesSeasons';
import { useActiveLibrary } from '../layout/useActiveLibrary';
import { useEffect } from 'react';
import { useLocation, useParams } from 'react-router-dom';
import { ErrorMessage, Loading } from '../components/common';
import { useResource } from '../lib/hooks';
import { duration, resolution } from '../lib/format';
import type { Movie, SeriesDetail } from '../lib/types';
import { MetadataEditor } from '../metadata/MetadataEditor';
import { Artwork, DetailHeader, PlayActions, TechnicalInfo } from './DetailParts';

export function MovieDetailPage() {
  const { id } = useParams();
  const { data, loading, error, reload } = useResource<{ item: Movie }>(`/api/media/${id}`);
  const item = data?.item;
  useActiveLibrary(item?.library_id);
  useEffect(() => {
    if (item) document.title = `${item.title} · Dinosaur`;
  }, [item]);
  return (
    <section className="movie-detail-page">
      <DetailHeader />
      {loading && !item && <Loading />}
      <ErrorMessage error={error} />
      {error && (
        <Button variant="outlined" onClick={reload}>
          重新加载
        </Button>
      )}
      {item && (
        <Paper className="detail-layout">
          <Artwork url={item.poster_url} title={item.title} />
          <article className="detail-copy">
            <h1>{item.title}</h1>
            <p className="detail-meta">
              {item.year || '年份未知'} · {duration(item.duration)} · {resolution(item)}
            </p>
            <PlayActions item={item} reload={reload} />
            <MetadataEditor
              key={item.id}
              id={item.id}
              kind="movie"
              year={item.year}
              info={item.metadata}
              reload={reload}
            />
            {item.probe_error && <p className="form-error">视频信息读取失败</p>}
            <TechnicalInfo item={item} />
          </article>
        </Paper>
      )}
    </section>
  );
}

export function SeriesDetailPage() {
  const location = useLocation();
  const { id } = useParams();
  const { data, loading, error, reload } = useResource<SeriesDetail>(`/api/series/${id}`);
  const item = data?.series;
  useActiveLibrary(item?.library_id);
  useEffect(() => {
    if (!data) return;
    document.title = `${data.series.title} · Dinosaur`;
  }, [data]);
  return (
    <section className="series-detail-page">
      <DetailHeader series />
      {loading && !item && <Loading />}
      <ErrorMessage error={error} />
      {error && (
        <Button variant="outlined" onClick={reload}>
          重新加载
        </Button>
      )}
      {item && data && (
        <Paper className="detail-layout">
          <Artwork url={item.poster_url} title={item.title} series />
          <article className="detail-copy">
            <h1>{item.title}</h1>
            <p className="detail-meta">
              {item.year || '年份未知'} · {item.season_count} 季 · {item.episode_count} 个文件 · 已看{' '}
              {item.watched_count}
            </p>
            {data.resume && (
              <div className="movie-play-actions">
                <ButtonLink
                  variant="contained"
                  to={`/watch/${data.resume.id}`}
                  state={{ kind: 'tv', from: location.pathname + location.search, backState: location.state }}
                >
                  {data.resume.position > 0 && !data.resume.watched ? '继续观看' : '播放'} ·{' '}
                  {data.resume.title}
                </ButtonLink>
              </div>
            )}
            <MetadataEditor
              key={item.id}
              id={item.id}
              kind="tv"
              year={item.year}
              info={item.metadata}
              reload={reload}
            />
            <SeriesSeasons data={data} reload={reload} />
          </article>
        </Paper>
      )}
    </section>
  );
}
