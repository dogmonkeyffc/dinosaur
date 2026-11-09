import { Button, Input } from '@fluentui/react-components';
import { useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { ErrorMessage, Icon, Loading, Pagination } from '../components/common';
import { useLibraries } from '../libraries/LibraryProvider';
import { useDebounced, useResource } from '../lib/hooks';
import { activeJob } from '../lib/format';
import type { MediaKind, Movie, Page, Series } from '../lib/types';
import { useJobs } from '../metadata/useJobs';
import { MediaCard } from './MediaCard';

export function CatalogPage({ kind }: { kind: MediaKind }) {
  const tv = kind === 'tv';
  const { libraries } = useLibraries();
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
  useJobs(kind, result.reload);
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
    <section id={tv ? 'series-section' : 'movies-section'} aria-label={tv ? '我的剧集' : '我的电影'}>
      <div className="movie-toolbar">
        <label className="search-field">
          <Icon name="search" />
          <Input
            className="form-control"
            type="search"
            maxLength={128}
            value={query}
            onChange={(event) => change('q', event.target.value)}
            placeholder={`搜索你的${tv ? '剧集' : '电影'}`}
            aria-label={tv ? '搜索剧集' : '搜索电影'}
          />
        </label>
      </div>
      <ErrorMessage error={result.error} />
      {result.loading && !result.data && <Loading />}
      <div className="movie-grid" aria-live="polite" aria-busy={result.loading}>
        {result.data?.items.map((item) => (
          <MediaCard key={item.id} item={item} />
        ))}
      </div>
      {!result.loading && !result.error && !result.data?.items.length && (
        <div className="empty-state">
          <h2>{query || library ? '没有找到匹配的内容' : '暂无媒体'}</h2>
          <Button
            className="button secondary"
            onClick={() => {
              result.reload();
            }}
          >
            刷新列表
          </Button>
        </div>
      )}
      {result.error && (
        <Button className="button secondary" onClick={result.reload}>
          重新加载
        </Button>
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
