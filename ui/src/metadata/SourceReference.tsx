import { InfoTip } from '../components/InfoTip';
import type { MediaKind } from '../lib/types';

export function SourceReference({
  item,
  kind = 'movie',
}: {
  item?: { tmdb_id: number; bangumi_id?: number; source?: string; source_url?: string };
  kind?: MediaKind;
}) {
  if (item?.source === 'agent' || item?.source === 'nfo') {
    const name = item.source === 'agent' ? 'Agent' : 'NFO';
    return (
      <span className="field-label">
        {item.source_url ? (
          <a href={item.source_url} target="_blank" rel="noopener noreferrer">
            {name}
          </a>
        ) : (
          name
        )}
        <InfoTip label="资料管理方式">资料已确认，自动刮削不会覆盖；通过 Agent API 更新。</InfoTip>
      </span>
    );
  }
  if (!item || (!item.tmdb_id && !item.bangumi_id)) return <>—</>;
  const bangumi = !!item.bangumi_id;
  const id = bangumi ? item.bangumi_id : item.tmdb_id;
  const href = bangumi
    ? `https://bgm.tv/subject/${id}`
    : `https://www.themoviedb.org/${kind === 'tv' ? 'tv' : 'movie'}/${id}`;
  return (
    <a href={href} target="_blank" rel="noopener noreferrer">
      {bangumi ? 'Bangumi' : 'TMDB'} · {id}
    </a>
  );
}
