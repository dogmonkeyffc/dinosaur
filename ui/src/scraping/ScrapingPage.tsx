import './records.css';
import { FormInput } from '../components/FormInput';
import { Paper, Table, TableContainer, TableHead, TableBody, TableRow, TableCell } from '@mui/material';
import { SourceReference } from '../metadata/SourceReference';
import { Select } from '../components/Select';
import { useState } from 'react';
import { ErrorMessage, Heading, Loading, Pagination } from '../components/common';
import { useLibraries } from '../libraries/LibraryProvider';
import { useDebounced, useResource } from '../lib/hooks';
import { scrapeMessage, stateNames, time } from '../lib/format';
import type { Page, ScrapeRecord } from '../lib/types';
import { StatusBadge } from './StatusBadge';
import { RecordActions, RecordCandidates } from './RecordActions';
import { RecordMedia } from './RecordMedia';
import { useAuth } from '../auth/AuthProvider';
import { scrapingStates, useScrapingFilters } from './useScrapingFilters';

export function ScrapingPage() {
  const { user } = useAuth();
  const { library, status, setLibrary, setStatus } = useScrapingFilters(user!.id);
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
      <Heading title="刮削管理" />
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
            {scrapingStates.map((value) => (
              <option key={value} value={value}>
                {value ? stateNames[value] : '全部状态'}
              </option>
            ))}
          </Select>
        </label>
        <label>
          搜索媒体
          <FormInput
            className="form-control"
            type="search"
            maxLength={128}
            value={query}
            onChange={(event) => filter(setQuery, event.target.value)}
            placeholder="片名、别名或文件名"
          />
        </label>
      </form>
      {resource.loading && !resource.data && <Loading />}
      <TableContainer
        component={Paper}
        className="scrape-table-wrap"
        role="region"
        aria-label="媒体识别结果"
        tabIndex={0}
      >
        <Table className="scrape-table" role="table">
          <TableHead role="rowgroup">
            <TableRow role="row">
              <TableCell role="columnheader" scope="col">
                媒体 / 文件
              </TableCell>
              <TableCell role="columnheader" scope="col">
                媒体库
              </TableCell>
              <TableCell role="columnheader" scope="col">
                状态 / 提示
              </TableCell>
              <TableCell role="columnheader" scope="col">
                资料来源 / 语言
              </TableCell>
              <TableCell role="columnheader" scope="col">
                更新时间
              </TableCell>
              <TableCell role="columnheader" scope="col">
                操作
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody role="rowgroup">
            {resource.data?.items.map((item) => (
              <TableRow role="row" key={item.series_id ? `tv:${item.series_id}` : `movie:${item.media_id}`}>
                <TableCell role="cell">
                  <RecordMedia item={item} />
                </TableCell>
                <TableCell role="cell">
                  <span className="scrape-cell-label" aria-hidden="true">
                    媒体库
                  </span>
                  {item.library_name}
                </TableCell>
                <TableCell role="cell">
                  <span className="scrape-cell-label" aria-hidden="true">
                    状态
                  </span>
                  <StatusBadge status={item.status} />
                  <p className="scrape-row-note">{scrapeMessage(item.message)}</p>
                  <RecordCandidates item={item} reload={resource.reload} />
                </TableCell>
                <TableCell role="cell">
                  <span className="scrape-cell-label" aria-hidden="true">
                    资料来源
                  </span>
                  <SourceReference item={item} kind={item.series_id ? 'tv' : 'movie'} />
                  <small>{item.language}</small>
                </TableCell>
                <TableCell role="cell">
                  <span className="scrape-cell-label" aria-hidden="true">
                    更新时间
                  </span>
                  {time(item.updated_at)}
                </TableCell>
                <TableCell role="cell">
                  <span className="scrape-cell-label" aria-hidden="true">
                    操作
                  </span>
                  <RecordActions item={item} reload={resource.reload} />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      {!resource.loading && resource.data && !resource.data.total && (
        <p className="scrape-admin-empty">没有符合筛选条件的媒体。</p>
      )}
      <Pagination page={page} total={resource.data?.total || 0} size={20} onChange={setPage} />
    </section>
  );
}
