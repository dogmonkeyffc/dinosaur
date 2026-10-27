import { useEffect, useState } from 'react';
import { Text, View } from 'react-native';
import type { Client } from '@dinosaur/client';
import { activeCue, parseVTT, type Cue } from '@dinosaur/playback-core/src/subtitles';

export function Subtitles({
  client,
  id,
  track,
  time,
  failed,
}: {
  client: Client;
  id: number;
  track?: number;
  time: number;
  failed: (error: string) => void;
}) {
  const [cues, setCues] = useState<Cue[]>([]);
  useEffect(() => {
    const controller = new AbortController();
    setCues([]);
    if (track === undefined) return;
    client
      .text(`/api/playback/${id}/subtitles/${track}`, controller.signal)
      .then((data) => {
        if (!controller.signal.aborted) setCues(parseVTT(data));
      })
      .catch(() => {
        if (!controller.signal.aborted) failed('字幕加载失败。');
      });
    return () => controller.abort();
  }, [client, id, track]);
  const text = activeCue(cues, time);
  return text ? (
    <View
      pointerEvents="none"
      style={{ position: 'absolute', bottom: 64, left: 24, right: 24, alignItems: 'center' }}
    >
      <Text
        style={{
          color: '#fff',
          backgroundColor: '#000b',
          padding: 8,
          textAlign: 'center',
          fontSize: 20,
          lineHeight: 28,
        }}
      >
        {text}
      </Text>
    </View>
  ) : null;
}
