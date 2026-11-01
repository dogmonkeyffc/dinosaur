import { Link } from 'react-router-dom';
import { Icon } from '../components/common';
import { Poster } from '../catalog/MediaCard';
import type { ScrapeRecord } from '../lib/types';

export function RecordMedia({ item }: { item: ScrapeRecord }) {
  const path = item.series_id ? `/series/${item.series_id}` : `/movies/${item.media_id}`;
  return (
    <div className="scrape-media">
      <Link to={path} className="scrape-media-cover" aria-label={`查看 ${item.title}`}>
        <Icon name={item.series_id ? 'tv' : 'film'} />
        <Poster url={item.poster_url} title={item.title} className="scrape-media-poster" />
      </Link>
      <div className="scrape-media-text">
        <strong>{item.title}</strong>
        <small>{item.filename}</small>
      </div>
    </div>
  );
}
