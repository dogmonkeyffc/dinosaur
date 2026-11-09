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
    return item.source_url ? (
      <a href={item.source_url} target="_blank" rel="noopener noreferrer">
        {name}
      </a>
    ) : (
      name
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
