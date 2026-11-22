import { useRef } from 'react';
import { Text, View } from 'react-native';
import type { Client } from '@dinosaur/client';
import type { MediaKind } from '@dinosaur/contracts';
import { Busy, Button, ErrorText, Screen } from '../components';
import { useMediaDetail } from '../catalog';
import { Poster } from '../Poster';
import { styles } from '../theme';

export function SeriesDetails({ client, id, kind, tv, active, season, lastSeason,
  width, close, selectSeason, play }: {
  client: Client;
  id: number;
  kind: MediaKind;
  tv: boolean;
  active: boolean;
  season?: number;
  lastSeason?: number;
  width: number;
  close: () => void;
  selectSeason: (value: number) => void;
  play: (id: number) => void;
}) {
  const { movie, series, error } = useMediaDetail(client, kind, id, active);
  const lastEpisode = useRef<number | undefined>(undefined);
  const scroll = useRef<Record<string, number>>({});
  const scrollKey = season === undefined ? 'seasons' : `season:${season}`;
  const item = movie || series?.series;
  const selected = series?.seasons.find((value) => value.number === season);
  const seasons = (series?.seasons || []).filter((value) =>
    Number.isSafeInteger(value.number) && value.number >= 0);
  const start = (episode: number) => {
    lastEpisode.current = episode;
    play(episode);
  };
  return (
    <Screen key={`${active}:${season}`} tv={tv}
      title={selected ? `${item?.title || ''} · ${season === 0 ? '特别篇' : `第 ${season} 季`}` : item?.title || '媒体详情'}
      onBack={close}
      backPreferred={tv && active && (!item || (series && season === undefined && !seasons.length)
        || (season !== undefined && !selected?.episodes.length))}
      scrollY={scroll.current[scrollKey] || 0}
      scrolled={(value) => { scroll.current[scrollKey] = value; }}>
      <ErrorText error={error} />
      {!item && !error && <Busy />}
      {item && season === undefined && (
        <View style={styles.card}>
          {!!item.year && <Text style={styles.muted}>{item.year}</Text>}
          {!!item.metadata?.overview && <Text style={styles.text}>{item.metadata.overview}</Text>}
          {movie && <Button primary preferred={tv && active} onPress={() => start(movie.id)}>
            {movie.position > 0 && !movie.watched ? '继续观看' : '播放'}
          </Button>}
        </View>
      )}
      {series && season === undefined && (
        <>
          {series.resume && (
            <Button primary onPress={() => start(series.resume!.id)}>
              {series.resume.position > 0 && !series.resume.watched ? '继续观看' : '播放'} · {series.resume.title}
            </Button>
          )}
          <View style={[styles.row, { alignItems: 'flex-start' }]}>
            {seasons.map((value, index) => (
              <Poster key={`${series.series.id}:${value.number}`} client={client} width={width}
                item={{ id: value.number, title: value.number === 0 ? '特别篇' : `第 ${value.number} 季`,
                  poster_url: value.episodes.find((episode) => episode.poster_url)?.poster_url || item?.poster_url }}
                preferred={tv && active && (lastSeason === value.number || (lastSeason === undefined && index === 0))}
                press={() => {
                  lastEpisode.current = undefined;
                  selectSeason(value.number);
                }} />
            ))}
          </View>
          {!seasons.length && <Text style={styles.muted}>暂无季度</Text>}
        </>
      )}
      {series && season !== undefined && (
        <View style={{ gap: 12 }}>
          {selected?.episodes.map((episode, index) => (
            <Button key={episode.id}
              preferred={tv && active && (lastEpisode.current === episode.id || (!lastEpisode.current && index === 0))}
              onPress={() => start(episode.id)}>
              {episode.watched ? '✓ ' : ''}{episode.title}
            </Button>
          ))}
          {!selected?.episodes.length && <Text style={styles.muted}>暂无分集</Text>}
        </View>
      )}
    </Screen>
  );
}
