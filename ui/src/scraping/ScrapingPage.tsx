import { SourceReference } from '../metadata/SourceReference';
import { Select } from '../components/Select';
import { useState } from 'react';
import { ErrorMessage, Loading, Pagination } from '../components/common';
import { useLibraries } from '../libraries/LibraryProvider';
import { useDebounced, useResource } from '../lib/hooks';
import { scrapeMessage, stateNames, time } from '../lib/format';
import type { Page, ScrapeRecord } from '../lib/types';
import { StatusBadge } from './StatusBadge';
import { RecordActions, RecordCandidates } from './RecordActions';

const states = ['', 'matched', 'unmatched', 'needs_review', 'failed', 'warning'];
export function ScrapingPage() {
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
  const resource = useResource<Page<ScrapeRecord>>('/api/admin/metadata/records?' + params, 3000);
  function filter(setter: (value: string) => void, value: string) {
    setter(value);
    setPage(1);
  }
  return (
    <section id="scraping-section" aria-label="刮削管理">
      <ErrorMessage error={resource.error} />
      <form className="scrape-admin-filters" onSubmit={(event) => event.preventDefault()}>
        <label>
          媒体库
          <Select value={library} onValueChange={(value) => filter(setLibrary, value)}>
            <option value="">全部媒体库</option>
            {libraries.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </Select>
        </label>
        <label>
          识别状态
          <Select value={status} onValueChange={(value) => filter(setStatus, value)}>
            {states.map((value) => (
              <option key={value} value={value}>
                {value ? stateNames[value] : '全部状态'}
              </option>
            ))}
          </Select>
        </label>
        <label>
          搜索媒体
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
      <div className="scrape-table-wrap" role="region" aria-label="媒体识别结果" tabIndex={0}>
        <table className="scrape-table">
          <thead>
            <tr>
              <th>媒体 / 文件</th>
              <th>媒体库</th>
              <th>状态 / 提示</th>
              <th>资料来源 / 语言</th>
              <th>更新时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {resource.data?.items.map((item) => (
              <tr key={item.series_id ? `tv:${item.series_id}` : `movie:${item.media_id}`}>
                <td>
                  <strong>{item.title}</strong>
                  <small>{item.filename}</small>
                </td>
                <td>{item.library_name}</td>
                <td>
                  <StatusBadge status={item.status} />
                  <p className="scrape-row-note">{scrapeMessage(item.message)}</p>
                  <RecordCandidates item={item} reload={resource.reload} />
                </td>
                <td>
                  <SourceReference item={item} kind={item.series_id ? 'tv' : 'movie'} />
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
        <p className="scrape-admin-empty">没有符合筛选条件的媒体。</p>
      )}
      <Pagination page={page} total={resource.data?.total || 0} size={20} onChange={setPage} />
    </section>
  );
}
