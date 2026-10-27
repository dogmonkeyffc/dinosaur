import { useState } from 'react';
import { Text, View } from 'react-native';
import type { Client } from '@dinosaur/client';
import type { MediaKind } from '@dinosaur/contracts';
import { Busy, Button, ErrorText, Poster, Screen, styles, useMediaDetail } from '@dinosaur/native-kit';

export function Details({
  client,
  kind,
  id,
  close,
  play,
}: {
  client: Client;
  kind: MediaKind;
  id: number;
  close: () => void;
  play: (id: number) => void;
}) {
  const { movie, series, error } = useMediaDetail(client, kind, id);
  const [season, setSeason] = useState<number>();
  const item = movie || series?.series;
  const selected = series?.seasons.find((value) => value.number === season) || series?.seasons[0];
  return (
    <Screen tv title={item?.title || '媒体详情'}>
      <Button onPress={close}>返回列表</Button>
      <ErrorText error={error} />
      {!item && !error && <Busy />}
      {item && (
        <View style={[styles.row, { alignItems: 'flex-start', gap: 32 }]}>
          <Poster
            client={client}
            item={item}
            width={180}
            press={() => {
              if (movie) play(movie.id);
              else if (series?.resume) play(series.resume.id);
            }}
          />
          <View style={[styles.card, { flex: 1 }]}>
            <Text style={styles.muted}>{item.year || ''}</Text>
            {item.metadata?.overview && (
              <Text numberOfLines={6} style={styles.text}>
                {item.metadata.overview}
              </Text>
            )}
            {movie && (
              <Button primary preferred onPress={() => play(movie.id)}>
                {movie.position > 0 && !movie.watched ? '继续观看' : '播放'}
              </Button>
            )}
            {series?.resume && (
              <Button primary preferred onPress={() => play(series.resume!.id)}>
                继续观看 · {series.resume.title}
              </Button>
            )}
          </View>
        </View>
      )}
      {series && (
        <>
          <View style={styles.row}>
            {series.seasons.map((value) => (
              <Button
                key={value.number}
                primary={selected?.number === value.number}
                onPress={() => setSeason(value.number)}
              >
                第 {value.number} 季
              </Button>
            ))}
          </View>
          <View style={styles.row}>
            {selected?.episodes.map((episode) => (
              <Button key={episode.id} onPress={() => play(episode.id)}>
                {episode.watched ? '✓ ' : ''}
                {episode.title}
              </Button>
            ))}
          </View>
        </>
      )}
    </Screen>
  );
}
