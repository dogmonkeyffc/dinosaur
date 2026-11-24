import { Rating } from '@mui/material';
import { Poster } from '../catalog/MediaCard';
import { Icon } from '../components/common';
import type { HistoryEntry as Entry } from './types';

export function HistoryEntry({ item }: { item: Entry }) {
  const douban = item.source === 'douban' && /^[1-9][0-9]*$/.test(item.source_id);
  return (
    <li className="watch-history-entry">
      <time className="watch-history-day" dateTime={item.watched_on} aria-label={item.watched_on}>
        {item.watched_on.slice(8)}
      </time>
      <article className="watch-history-card" aria-labelledby={`history-title-${item.id}`}>
        <div className="watch-history-artwork">
          <Icon name="film" />
          <Poster
            key={item.poster_url}
            url={item.poster_url}
            title={item.title}
            className="watch-history-poster"
          />
        </div>
        <div className="watch-history-info">
          <h3 id={`history-title-${item.id}`}>{item.title}</h3>
          {!!item.original_title && <p className="watch-history-original muted">{item.original_title}</p>}
          <div className="watch-history-meta">
            {item.year > 0 && <span className="muted">{item.year}</span>}
            {item.rating > 0 && (
              <Rating value={item.rating} readOnly size="small" aria-label={`${item.rating} 星`} />
            )}
            {douban ? (
              <a
                href={`https://movie.douban.com/subject/${item.source_id}/`}
                target="_blank"
                rel="noopener noreferrer"
              >
                豆瓣
              </a>
            ) : (
              <span className="muted">{item.source === 'local' ? '站内' : '手动记录'}</span>
            )}
          </div>
          {!!item.review && <p className="watch-history-review">{item.review}</p>}
        </div>
      </article>
    </li>
  );
}
