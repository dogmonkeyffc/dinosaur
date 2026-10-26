import { useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useAuth } from '../auth/AuthProvider';
import { ErrorMessage, Heading, Icon, Loading, Pagination } from '../components/common';
import { useLibraries } from '../libraries/LibraryProvider';
import { useAction, useDebounced, useResource } from '../lib/hooks';
import { activeJob, jobDescription, scanDescription } from '../lib/format';
import type { MediaKind, Movie, Page, Series } from '../lib/types';
import { useJobs } from '../metadata/useJobs';
import { MediaCard } from './MediaCard';

export function CatalogPage({ kind }: { kind: MediaKind }) {
  const tv = kind === 'tv';
  const { user } = useAuth();
  const { libraries, scan } = useLibraries();
  const [params, setParams] = useSearchParams();
  const page = Math.max(1, Number(params.get('page')) || 1);
  const library = params.get('library_id') || '';
  const query = params.get('q') || '';
  const search = useDebounced(query);
  const choices = libraries.filter((item) => item.type === kind);
  const scans = choices.filter((item) => activeJob(item.scan?.status));
  const scanSignature = choices
    .map((item) => `${item.id}:${item.scan?.status}:${item.scan?.processed}`)
    .join(',');
  const apiQuery = new URLSearchParams({ page: String(page), page_size: '24' });
  if (library) apiQuery.set('library_id', library);
  if (search.trim()) apiQuery.set('q', search.trim());
  const result = useResource<Page<Movie | Series>>(
    `/api/${tv ? 'series' : 'media'}?${apiQuery}`,
    scans.length ? 3000 : 0,
  );
  const recent = useResource<{ items: Movie[] }>(`/api/playback/continue?type=${kind}`);
  const jobs = useJobs(kind, result.reload);
  const action = useAction();
  useEffect(() => {
    result.reload();
  }, [scanSignature, result.reload]);
  useEffect(() => {
    if (result.data?.total && page > Math.ceil(result.data.total / 24)) change('page', '1');
  }, [result.data, page]);
  function change(name: string, value: string) {
    const next = new URLSearchParams(params);
    if (value) next.set(name, value);
    else next.delete(name);
    if (name !== 'page') next.delete('page');
    setParams(next, { replace: name === 'q' });
  }
  return (
    <section id={tv ? 'series-section' : 'movies-section'}>
      <Heading
        title={
          <>
            我的{tv ? '剧集' : '电影'}
            <span className="collection-count">{result.data?.total || 0}</span>
          </>
        }
      >
        {user?.role === 'admin' && (
          <div className="metadata-actions">
            <button
              className="button secondary"
              disabled={!choices.length || action.busy}
              onClick={() =>
                action.run(() => scan(choices.filter((item) => !library || String(item.id) === library)))
              }
            >
              扫描媒体库
            </button>
            <button
              className="button secondary"
              disabled={jobs.busy || jobs.running}
              onClick={() => jobs.start({ library_id: Number(library) })}
            >
              刮削未识别{tv ? '剧集' : '影片'}
            </button>
          </div>
        )}
      </Heading>
      {scans.length > 0 && (
        <div className="info-banner" role="status">
          {scans.map((item) => `${item.name}：${scanDescription(item.scan!)}`).join('；')}
        </div>
      )}
      {jobs.latest && (
        <p className="scrape-status" role="status">
          {jobDescription(jobs.latest)}
        </p>
      )}
      {(recent.data?.items.length || 0) > 0 && (
        <section className="continue-section">
          <h2>继续观看</h2>
          <div className="movie-grid continue-grid">
            {recent.data!.items.map((item) => (
              <MediaCard key={item.id} item={item} resume />
            ))}
          </div>
        </section>
      )}
      <div className="movie-toolbar">
        <label className="library-select">
          <span className="sr-only">选择媒体库</span>
          <select value={library} onChange={(event) => change('library_id', event.target.value)}>
            <option value="">全部{tv ? '剧集' : ''}媒体库</option>
            {choices.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </select>
        </label>
        <label className="search-field">
          <Icon name="search" />
          <input
            type="search"
            maxLength={128}
            value={query}
            onChange={(event) => change('q', event.target.value)}
            placeholder={`搜索你的${tv ? '剧集' : '电影'}`}
            aria-label={tv ? '搜索剧集' : '搜索电影'}
          />
        </label>
      </div>
      <ErrorMessage error={result.error || jobs.error || action.error || recent.error} />
      {result.loading && !result.data && <Loading />}
      <div className="movie-grid" aria-live="polite" aria-busy={result.loading}>
        {result.data?.items.map((item) => (
          <MediaCard key={item.id} item={item} />
        ))}
      </div>
      {!result.loading && !result.error && !result.data?.items.length && (
        <div className="empty-state">
          <h2>{query || library ? '没有找到匹配的内容' : '珍藏，即将登场'}</h2>
          <button
            className="button secondary"
            onClick={() => {
              result.reload();
              recent.reload();
            }}
          >
            刷新列表
          </button>
        </div>
      )}
      {result.error && (
        <button className="button secondary" onClick={result.reload}>
          重新加载
        </button>
      )}
      <Pagination
        page={page}
        total={result.data?.total || 0}
        size={24}
        onChange={(value) => change('page', String(value))}
      />
    </section>
  );
}
