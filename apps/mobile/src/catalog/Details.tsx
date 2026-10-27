import { useState } from 'react';
import { Text, View } from 'react-native';
import type { Client } from '@dinosaur/client';
import type { MediaKind } from '@dinosaur/contracts';
import { Busy, Button, ErrorText, Screen, styles, useMediaDetail } from '@dinosaur/native-kit';

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
    <Screen title={item?.title || '媒体详情'}>
      <Button onPress={close}>返回</Button>
      <ErrorText error={error} />
      {!item && !error && <Busy />}
      {item && (
        <View style={styles.card}>
          <Text style={styles.muted}>{item.year || ''}</Text>
          {item.metadata?.overview && <Text style={styles.text}>{item.metadata.overview}</Text>}
          {movie && (
            <Button primary onPress={() => play(movie.id)}>
              {movie.position > 0 && !movie.watched ? '继续观看' : '播放'}
            </Button>
          )}
          {series?.resume && (
            <Button primary onPress={() => play(series.resume!.id)}>
              继续观看 · {series.resume.title}
            </Button>
          )}
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
          {selected?.episodes.map((episode) => (
            <Button key={episode.id} onPress={() => play(episode.id)}>
              {episode.watched ? '✓ ' : ''}
              {episode.title}
            </Button>
          ))}
        </>
      )}
    </Screen>
  );
}
