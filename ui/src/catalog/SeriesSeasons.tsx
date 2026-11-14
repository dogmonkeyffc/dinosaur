import { Button, Card, CardActionArea } from '@mui/material';
import { Link, useLocation, useSearchParams } from 'react-router-dom';
import { Icon } from '../components/common';
import type { SeriesDetail } from '../lib/types';
import { EpisodeCard } from './DetailParts';
import { Poster } from './MediaCard';

function seasonTitle(number: number) {
  return number === 0 ? '特别篇' : `第 ${number} 季`;
}

export function SeriesSeasons({ data, reload }: { data: SeriesDetail; reload: () => void }) {
  const location = useLocation();
  const [params] = useSearchParams();
  const value = params.get('season');
  const selected =
    value !== null && /^\d+$/.test(value)
      ? data.seasons.find((season) => season.number === Number(value))
      : undefined;
  function seasonPath(number?: number) {
    const next = new URLSearchParams(params);
    if (number === undefined) next.delete('season');
    else next.set('season', String(number));
    return { search: next.size ? `?${next}` : '' };
  }
  return (
    <section className="series-seasons" aria-label="季度与分集">
      {selected ? (
        <>
          <div className="shelf-heading">
            <h2>{seasonTitle(selected.number)}</h2>
            <Button component={Link} to={seasonPath()} state={location.state}>
              返回季度列表
            </Button>
          </div>
          <div className="series-episodes">
            {selected.episodes.map((episode) => (
              <EpisodeCard key={episode.id} item={episode} reload={reload} />
            ))}
          </div>
        </>
      ) : (
        <>
          <h2>季度</h2>
          <div className="movie-grid">
            {data.seasons.map((season) => {
              const title = seasonTitle(season.number);
              return (
                <Card variant="outlined" key={season.number}>
                  <CardActionArea
                    component={Link}
                    to={seasonPath(season.number)}
                    state={location.state}
                    className="movie-card"
                    aria-label={`查看${title}`}
                  >
                    <span className="movie-artwork">
                      <span className="movie-placeholder">
                        <Icon name="tv" />
                      </span>
                      <Poster url={data.series.poster_url} title={title} className="movie-poster" />
                    </span>
                    <span className="movie-title">{title}</span>
                    <span className="movie-meta">{season.episodes.length} 集</span>
                  </CardActionArea>
                </Card>
              );
            })}
          </div>
          {data.seasons.length === 0 && <p>暂无季度</p>}
        </>
      )}
    </section>
  );
}
