import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { ErrorMessage, Loading } from '../components/common';
import { useResource } from '../lib/hooks';
import { duration, resolution } from '../lib/format';
import type { Movie, SeriesDetail } from '../lib/types';
import { MetadataEditor } from '../metadata/MetadataEditor';
import { Artwork, DetailHeader, EpisodeCard, PlayActions, TechnicalInfo } from './DetailParts';

export function MovieDetailPage() {
  const { id } = useParams();
  const { data, loading, error, reload } = useResource<{ item: Movie }>(`/api/media/${id}`);
  const item = data?.item;
  useEffect(() => {
    if (item) document.title = `${item.title} · Dinosaur`;
  }, [item]);
  return (
    <section className="movie-detail-page">
      <DetailHeader />
      {loading && !item && <Loading />}
      <ErrorMessage error={error} />
      {error && (
        <button className="button secondary detail-retry" onClick={reload}>
          重新加载
        </button>
      )}
      {item && (
        <div className="detail-layout">
          <Artwork url={item.poster_url} title={item.title} />
          <article className="detail-copy">
            <span className="small-label">{item.library_name}</span>
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
            {item.probe_error && (
              <p className="form-error">视频信息暂未读取完整，可以重新扫描。电影文件已被索引。</p>
            )}
            <TechnicalInfo item={item} />
          </article>
        </div>
      )}
    </section>
  );
}

export function SeriesDetailPage() {
  const { id } = useParams();
  const { data, loading, error, reload } = useResource<SeriesDetail>(`/api/series/${id}`);
  const [season, setSeason] = useState<number | null>(null);
  const item = data?.series;
  useEffect(() => {
    if (!data) return;
    document.title = `${data.series.title} · Dinosaur`;
    setSeason((previous) =>
      data.seasons.some((item) => item.number === previous)
        ? previous
        : (data.resume?.season_number ?? data.seasons[0]?.number ?? null),
    );
  }, [data]);
  return (
    <section className="series-detail-page">
      <DetailHeader series />
      {loading && !item && <Loading />}
      <ErrorMessage error={error} />
      {error && (
        <button className="button secondary detail-retry" onClick={reload}>
          重新加载
        </button>
      )}
      {item && data && (
        <div className="detail-layout">
          <Artwork url={item.poster_url} title={item.title} series />
          <article className="detail-copy">
            <span className="small-label">{item.library_name}</span>
            <h1>{item.title}</h1>
            <p className="detail-meta">
              {item.year || '年份未知'} · {item.season_count} 季 · {item.episode_count} 个文件 · 已看{' '}
              {item.watched_count}
            </p>
            {data.resume && (
              <div className="movie-play-actions">
                <Link className="button primary" to={`/watch/${data.resume.id}`} state={{ kind: 'tv' }}>
                  {data.resume.position > 0 && !data.resume.watched ? '继续观看' : '播放'} ·{' '}
                  {data.resume.title}
                </Link>
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
            <div className="series-seasons">
              <label>
                选择季度
                <select value={season ?? ''} onChange={(event) => setSeason(Number(event.target.value))}>
                  {data.seasons.map((item) => (
                    <option key={item.number} value={item.number}>
                      {item.number === 0 ? '特别篇' : `第 ${item.number} 季`} · {item.episodes.length} 个文件
                    </option>
                  ))}
                </select>
              </label>
              <div className="series-episodes">
                {data.seasons
                  .find((item) => item.number === season)
                  ?.episodes.map((episode) => (
                    <EpisodeCard key={episode.id} item={episode} reload={reload} />
                  ))}
              </div>
            </div>
          </article>
        </div>
      )}
    </section>
  );
}
