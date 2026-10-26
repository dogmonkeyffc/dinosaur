import { useState } from 'react';
import { Link } from 'react-router-dom';
import { ErrorMessage, Heading, Loading, Pagination } from '../components/common';
import { useLibraries } from '../libraries/LibraryProvider';
import { useDebounced, useResource } from '../lib/hooks';
import { scrapeMessage, stateNames, time } from '../lib/format';
import type { Page, ScrapeRecord } from '../lib/types';
import { useJobs } from '../metadata/useJobs';
import { Badge, JobsView } from './JobsView';
import { RecordActions, RecordCandidates } from './RecordActions';

const states = ['', 'matched', 'unmatched', 'needs_review', 'failed', 'warning'];
export function ScrapingPage() {
  const [tab, setTab] = useState('records');
  const [library, setLibrary] = useState('');
  const [status, setStatus] = useState('');
  const [query, setQuery] = useState('');
  const search = useDebounced(query);
  const [page, setPage] = useState(1);
  const { libraries } = useLibraries();
  const params = new URLSearchParams({ page: String(page), page_size: '20' });
  if (library) params.set('library_id', library);
  if (status) params.set('status', status);
  if (search.trim()) params.set('q', search.trim());
  const resource = useResource<Page<ScrapeRecord> & { summary: Record<string, number> }>(
    tab === 'records' ? '/api/admin/metadata/records?' + params : null,
    3000,
  );
  const jobs = useJobs('movie', resource.reload);
  function filter(setter: (value: string) => void, value: string) {
    setter(value);
    setPage(1);
  }
  return (
    <section id="scraping-section">
      <Heading title="刮削管理">
        <div className="metadata-actions">
          <Link className="text-button" to="/settings/metadata">
            TMDB 设置
          </Link>
          <button
            className="button secondary"
            onClick={() => {
              resource.reload();
              jobs.reload();
            }}
          >
            刷新状态
          </button>
          <button
            className="button primary"
            disabled={jobs.busy || jobs.running}
            onClick={() => jobs.start({ library_id: tab === 'records' ? Number(library) : 0 })}
          >
            刮削未识别影片
          </button>
        </div>
      </Heading>
      <div className="scrape-admin-tabs" role="tablist" aria-label="刮削管理视图">
        {[
          ['records', '影片状态'],
          ['jobs', '任务历史'],
        ].map(([value, label]) => (
          <button
            key={value}
            className="button secondary"
            role="tab"
            aria-selected={tab === value}
            onClick={() => setTab(value)}
          >
            {label}
          </button>
        ))}
      </div>
      <ErrorMessage error={resource.error || jobs.error} />
      {tab === 'jobs' ? (
        <JobsView />
      ) : (
        <>
          <div className="scrape-summary" aria-label="识别状态统计">
            {states.map((value) => (
              <button
                key={value}
                className="scrape-summary-card"
                aria-pressed={status === value}
                onClick={() => filter(setStatus, value)}
              >
                <strong>{resource.data?.summary[value || 'all'] || 0}</strong>
                <span>{value ? stateNames[value] : '全部影片'}</span>
              </button>
            ))}
          </div>
          <form className="scrape-admin-filters" onSubmit={(event) => event.preventDefault()}>
            <label>
              媒体库
              <select value={library} onChange={(event) => filter(setLibrary, event.target.value)}>
                <option value="">全部媒体库</option>
                {libraries
                  .filter((item) => item.type === 'movie')
                  .map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name}
                    </option>
                  ))}
              </select>
            </label>
            <label>
              识别状态
              <select value={status} onChange={(event) => filter(setStatus, event.target.value)}>
                {states.map((value) => (
                  <option key={value} value={value}>
                    {value ? stateNames[value] : '全部状态'}
                  </option>
                ))}
              </select>
            </label>
            <label>
              搜索影片
              <input
                type="search"
                maxLength={128}
                value={query}
                onChange={(event) => filter(setQuery, event.target.value)}
                placeholder="片名、别名或文件名"
              />
            </label>
          </form>
          {resource.loading && !resource.data && <Loading />}
          <div className="scrape-table-wrap" role="region" aria-label="影片识别结果" tabIndex={0}>
            <table className="scrape-table">
              <thead>
                <tr>
                  <th>影片 / 文件</th>
                  <th>媒体库</th>
                  <th>状态 / 提示</th>
                  <th>TMDB / 语言</th>
                  <th>更新时间</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                {resource.data?.items.map((item) => (
                  <tr key={item.media_id}>
                    <td>
                      <strong>{item.title}</strong>
                      <small>{item.filename}</small>
                    </td>
                    <td>{item.library_name}</td>
                    <td>
                      <Badge status={item.status} />
                      <p className="scrape-row-note">{scrapeMessage(item.message)}</p>
                      <RecordCandidates item={item} reload={resource.reload} />
                    </td>
                    <td>
                      {item.tmdb_id || '—'}
                      <small>{item.language}</small>
                    </td>
                    <td>{time(item.updated_at)}</td>
                    <td>
                      <RecordActions item={item} reload={resource.reload} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {!resource.loading && resource.data && !resource.data.total && (
            <p className="scrape-admin-empty">没有符合筛选条件的影片。</p>
          )}
          <Pagination page={page} total={resource.data?.total || 0} size={20} onChange={setPage} />
        </>
      )}
    </section>
  );
}
