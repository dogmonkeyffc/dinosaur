import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import { useAction, useResource } from '../lib/hooks';
import type { Candidate, ScrapeRecord } from '../lib/types';
import { useManualMatch } from '../metadata/useManualMatch';
import { CandidateList } from '../metadata/MetadataEditor';
import { ErrorMessage } from '../components/common';

export function RecordActions({ item, reload }: { item: ScrapeRecord; reload: () => void }) {
  const action = useAction();
  const id = item.series_id || item.media_id;
  const managed = item.source === 'agent' || item.source === 'nfo';
  if (!id) return <span className="muted">媒体已移除</span>;
  return (
    <div className="scrape-row-actions">
      <Link
        className="text-button"
        to={`/${item.series_id ? 'series' : 'movies'}/${id}`}
        state={{ from: '/settings/scraping' }}
      >
        {managed ? '查看详情' : '详情 / 匹配更正'}
      </Link>
      {!managed && (
        <button
          className="text-button"
          disabled={action.busy}
          onClick={() =>
            action.run(async () => {
              await api(`/api/admin/${item.series_id ? 'tv/' : ''}metadata/jobs`, {
                method: 'POST',
                body: {
                  [item.series_id ? 'series_id' : 'media_id']: id,
                  refresh: item.tmdb_id > 0 || !!item.bangumi_id,
                },
              });
              reload();
            })
          }
        >
          {item.tmdb_id > 0 || !!item.bangumi_id ? '刷新资料' : '重试识别'}
        </button>
      )}
      <ErrorMessage error={action.error} />
    </div>
  );
}

export function RecordCandidates({ item, reload }: { item: ScrapeRecord; reload: () => void }) {
  const ambiguous = item.message?.split(',').includes('ambiguous');
  const resource = useResource<{ candidates: Candidate[] }>(
    ambiguous && !item.candidates?.length
      ? `/api/admin/${item.series_id ? 'series' : 'media'}/${item.series_id || item.media_id}/metadata/candidates`
      : null,
  );
  const match = useManualMatch(item.series_id ? 'tv' : 'movie', item.series_id || item.media_id, reload);
  const candidates = resource.data?.candidates || item.candidates || [];
  if (!ambiguous && !candidates.length) return null;
  return (
    <section className="candidate-preview">
      <h3>相似候选（最多 5 部）</h3>
      <ErrorMessage error={resource.error || match.error} />
      {match.notice && <p role="status">{match.notice}</p>}
      <CandidateList candidates={candidates} disabled={match.busy} select={match.select} />
      {resource.error && (
        <button className="text-button" onClick={resource.reload}>
          重新加载候选
        </button>
      )}
    </section>
  );
}
