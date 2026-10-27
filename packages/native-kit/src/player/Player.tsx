import { useEffect, useState } from 'react';
import { View } from 'react-native';
import type { Client } from '@dinosaur/client';
import type { PlaybackInfo } from '@dinosaur/contracts';
import { Busy, Button, ErrorText } from '../components';
import { styles } from '../theme';
import { Stage } from './Stage';

export function Player({
  client,
  id,
  close,
  tv = false,
}: {
  client: Client;
  id: number;
  close: () => void;
  tv?: boolean;
}) {
  const [currentID, setCurrentID] = useState(id);
  const [mode, setMode] = useState('original');
  const [target, setTarget] = useState<number>();
  const [source, setSource] = useState<{ info: PlaybackInfo; start: number }>();
  const [error, setError] = useState('');
  useEffect(() => {
    let active = true;
    setSource(undefined);
    setError('');
    client
      .playback(currentID, mode, target || 0)
      .then((info) => {
        if (active) setSource({ info, start: target ?? info.resume_position });
      })
      .catch((reason) => {
        if (active) setError(reason instanceof Error ? reason.message : '无法开始播放。');
      });
    return () => {
      active = false;
    };
  }, [client, currentID, mode, target]);
  if (!source)
    return (
      <View style={styles.centered}>
        {!error && <Busy />}
        <ErrorText error={error} />
        <Button onPress={close}>返回</Button>
      </View>
    );
  return (
    <Stage
      key={`${currentID}:${mode}:${target}`}
      client={client}
      info={source.info}
      start={source.start}
      tv={tv}
      close={close}
      restart={(position, compatible) => {
        setSource(undefined);
        setMode(compatible ? 'compatible' : 'original');
        setTarget(position);
      }}
      next={() => {
        if (source.info.next_episode) {
          setSource(undefined);
          setMode('original');
          setTarget(undefined);
          setCurrentID(source.info.next_episode.id);
        }
      }}
    />
  );
}
