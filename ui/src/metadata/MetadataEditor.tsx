import { Select } from '../components/Select';
import { SourceReference } from './SourceReference';
import { InfoTip } from '../components/InfoTip';
import { useState, type FormEvent } from 'react';
import { useAuth } from '../auth/AuthProvider';
import { ErrorMessage } from '../components/common';
import { jobDescription, scrapeMessage } from '../lib/format';
import { useResource } from '../lib/hooks';
import type { Candidate, MediaKind, Metadata } from '../lib/types';
import { useJobs } from './useJobs';

export function CandidateList({
  candidates,
  select,
  disabled = false,
}: {
  candidates: Candidate[];
  select: (id: number, provider?: string) => void;
  disabled?: boolean;
}) {
  return (
    <div className="candidate-list">
      {candidates.slice(0, 5).map((item) => (
        <article key={`${item.provider || 'tmdb'}:${item.id}`} className="metadata-candidate">
          <strong>
            {item.title} ({item.year || '年份未知'})
          </strong>
          <span>{item.original_title}</span>
          <p>{item.overview}</p>
          <button
            className="button secondary"
            disabled={disabled}
            onClick={() => select(item.id, item.provider)}
          >
            {item.provider === 'bangumi' ? 'Bangumi' : 'TMDB'} · {item.id} · 选择
          </button>
        </article>
      ))}
    </div>
  );
}

export function MetadataEditor({
  id,
  kind,
  year,
  info,
  reload,
}: {
  id: number;
  kind: MediaKind;
  year: number;
  info?: Metadata;
  reload: () => void;
}) {
  const { user } = useAuth();
  const admin = user?.role === 'admin';
  const matched = !!(info?.tmdb_id || info?.bangumi_id);
  const [source, setSource] = useState(info?.bangumi_id ? 'bangumi' : 'tmdb');
  const endpoint = `/api/admin/${kind === 'tv' ? 'series' : 'media'}/${id}/metadata/candidates`;
  const [query, setQuery] = useState<string | null>(null);
  const ambiguous = !info?.candidates?.length && info?.message?.split(',').includes('ambiguous');
  const search = useResource<{ candidates: Candidate[] }>(
    admin ? (query ?? (ambiguous ? endpoint : null)) : null,
  );
  const jobs = useJobs(kind, reload);
  const candidates = search.data?.candidates || info?.candidates || [];
  const start = (providerID?: number, provider = 'tmdb') =>
    jobs.start({
      [kind === 'tv' ? 'series_id' : 'media_id']: id,
      refresh: matched && !providerID,
      ...(providerID ? { [provider === 'bangumi' ? 'bangumi_id' : 'tmdb_id']: providerID } : {}),
    });
  function find(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const values = new FormData(event.currentTarget);
    const params = new URLSearchParams();
    if (values.get('query')) params.set('q', String(values.get('query')).trim());
    if (values.get('year') !== '') params.set('year', String(values.get('year')));
    setQuery(endpoint + '?' + params);
    search.reload();
  }
  return (
    <>
      {info?.original_title && <p className="movie-original-title">{info.original_title}</p>}
      {info?.overview && <p className="movie-overview">{info.overview}</p>}
      <p className="movie-metadata-status">
        {matched
          ? '已匹配'
          : info?.status === 'needs_review'
            ? '需要确认匹配'
            : info?.status === 'failed'
              ? '刮削失败'
              : '尚未识别资料'}{' '}
        {matched && <SourceReference item={info} kind={kind} />}
      </p>
      {info?.message && <p className="field-help">{scrapeMessage(info.message)}</p>}
      {admin && (
        <>
          <div className="metadata-detail-actions">
            <button className="text-button" disabled={jobs.busy || jobs.running} onClick={() => start()}>
              {matched ? '刷新已匹配资料' : '自动识别'}
            </button>
          </div>
          {jobs.latest && jobs.running && (
            <p className="scrape-status" role="status">
              {jobDescription(jobs.latest)}
            </p>
          )}
          <ErrorMessage error={jobs.error} />
          <details
            className={kind === 'tv' ? 'series-matching' : 'movie-matching'}
            open={info?.status === 'needs_review' || undefined}
          >
            <summary>匹配 / 更正{kind === 'tv' ? '剧集' : '电影'}</summary>
            <form onSubmit={find}>
              <label htmlFor={`metadata-query-${kind}-${id}`}>
                <span className="field-label">
                  <span id={`metadata-query-label-${kind}-${id}`}>{kind === 'tv' ? '剧名' : '电影名'}</span>
                  <InfoTip label="候选搜索名称">留空使用本地名称。</InfoTip>
                </span>
                <input
                  aria-labelledby={`metadata-query-label-${kind}-${id}`}
                  id={`metadata-query-${kind}-${id}`}
                  name="query"
                  maxLength={128}
                />
              </label>
              <label>
                年份
                <input
                  name="year"
                  type="number"
                  min={0}
                  max={2100}
                  defaultValue={year || ''}
                  placeholder="不限"
                />
              </label>
              <button className="button secondary" disabled={search.loading}>
                搜索候选
              </button>
            </form>
            <form
              onSubmit={(event) => {
                event.preventDefault();
                start(Number(new FormData(event.currentTarget).get('provider_id')), source);
              }}
            >
              <label>
                资料来源
                <Select value={source} onValueChange={setSource}>
                  <option value="tmdb">TMDB</option>
                  <option value="bangumi">Bangumi</option>
                </Select>
              </label>
              <label htmlFor={`metadata-provider-${kind}-${id}`}>
                <span className="field-label">
                  <span id={`metadata-provider-label-${kind}-${id}`}>
                    {source === 'bangumi' ? 'Bangumi ID' : 'TMDB ID'}
                  </span>
                  <InfoTip label="条目 ID 位置">
                    {source === 'bangumi'
                      ? 'Bangumi 条目网址 subject/ 后的数字 ID。'
                      : 'TMDB 电影或剧集详情页网址中的数字 ID。'}
                  </InfoTip>
                </span>
                <input
                  aria-labelledby={`metadata-provider-label-${kind}-${id}`}
                  id={`metadata-provider-${kind}-${id}`}
                  name="provider_id"
                  type="number"
                  min={1}
                  max={2147483647}
                  required
                />
              </label>
              <button className="button secondary" disabled={jobs.busy || jobs.running}>
                使用此 ID 匹配
              </button>
            </form>
            <ErrorMessage error={search.error} />
            {search.loading && <p role="status">正在搜索候选…</p>}
            <CandidateList candidates={candidates} select={start} disabled={jobs.busy || jobs.running} />
            {search.data && !candidates.length && <p className="field-help">未找到候选</p>}
          </details>
        </>
      )}
    </>
  );
}
