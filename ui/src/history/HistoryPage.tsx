import { Button } from '@mui/material';
import { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { ErrorMessage, Heading, Loading, Pagination } from '../components/common';
import { useResource } from '../lib/hooks';
import type { Page } from '../lib/types';
import { HistoryEntry } from './HistoryEntry';
import type { HistoryEntry as Entry } from './types';
import './history.css';
import { HistoryRecordingSettings } from './HistoryRecordingSettings';

export function HistoryPage() {
  const [settings, setSettings] = useState(false);
  const [params, setParams] = useSearchParams();
  const requested = Number(params.get('page') ?? 1);
  const page = Number.isSafeInteger(requested) && requested > 0 && requested <= 1000000 ? requested : 1;
  const { data, loading, error, reload } = useResource<Page<Entry>>(`/api/history?page=${page}&page_size=24`);
  const groups = new Map<string, Entry[]>();
  for (const item of data?.items ?? []) {
    const month = item.watched_on.slice(0, 7);
    groups.set(month, [...(groups.get(month) ?? []), item]);
  }
  useEffect(() => {
    if (data && page > Math.max(1, Math.ceil(data.total / 24))) {
      const last = Math.max(1, Math.ceil(data.total / 24));
      setParams(last === 1 ? {} : { page: String(last) }, { replace: true });
    }
  }, [data, page, setParams]);
  function changePage(next: number) {
    setParams(next === 1 ? {} : { page: String(next) });
    window.scrollTo(0, 0);
  }
  return (
    <section className="watch-history-page" aria-busy={loading}>
      <Heading title="观看记录">
        <div className="watch-history-actions">
          <Button variant="outlined" onClick={() => setSettings(true)}>
            记录设置
          </Button>
          {data && (
            <span className="muted" role="status">
              {data.total} 条
            </span>
          )}
          <Button variant="outlined" disabled={loading} onClick={reload}>
            刷新
          </Button>
        </div>
      </Heading>
      <ErrorMessage error={error} />
      {!!error && <Button onClick={reload}>重试</Button>}
      {loading && !data && <Loading />}
      {!loading && !error && data?.total === 0 && (
        <p className="muted" role="status">
          暂无观看记录
        </p>
      )}
      <div className="watch-history-content">
        {[...groups].map(([month, items]) => (
          <section className="watch-history-month" key={month} aria-labelledby={`history-month-${month}`}>
            <h2 id={`history-month-${month}`}>
              {month.slice(0, 4)}年{Number(month.slice(5))}月
            </h2>
            <ol className="watch-history-list">
              {items.map((item) => (
                <HistoryEntry key={item.id} item={item} />
              ))}
            </ol>
          </section>
        ))}
      </div>
      {data && <Pagination page={page} total={data.total} size={24} onChange={changePage} />}
      {settings && <HistoryRecordingSettings close={() => setSettings(false)} />}
    </section>
  );
}
