import { useEffect, useRef, useState } from 'react';
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
  const [subtitle, setSubtitle] = useState<number | null>();
  const burnID = useRef<number | undefined>(undefined);
  const [revision, setRevision] = useState(0);
  const [mode, setMode] = useState('original');
  const [target, setTarget] = useState<number>();
  const [source, setSource] = useState<{ info: PlaybackInfo; start: number }>();
  const [error, setError] = useState('');
  useEffect(() => {
    let active = true;
    setSource(undefined);
    setError('');
    client
      .playback(
        currentID,
        mode,
        target || 0,
        undefined,
        burnID.current ?? (subtitle === undefined ? 'auto' : undefined),
      )
      .then((info) => {
        if (active) {
          if (subtitle === undefined) {
            const selected = info.selected_subtitle ?? null;
            setSubtitle(selected);
            burnID.current =
              info.subtitles.find((track) => track.id === selected)?.delivery === 'burn'
                ? (selected ?? undefined)
                : undefined;
          }
          setSource({ info, start: target ?? info.resume_position });
        }
      })
      .catch((reason) => {
        if (active) setError(reason instanceof Error ? reason.message : '无法开始播放。');
      });
    return () => {
      active = false;
    };
  }, [client, currentID, mode, target, revision]);
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
      track={subtitle ?? undefined}
      selectSubtitle={(id, position) => {
        const selected = source.info.subtitles.find((value) => value.id === id);
        setSubtitle(id ?? null);
        if (burnID.current !== undefined || selected?.delivery === 'burn') {
          setSource(undefined);
          setTarget(position);
          burnID.current = selected?.delivery === 'burn' ? id : undefined;
          setRevision((value) => value + 1);
        }
      }}
      close={close}
      restart={(position, compatible) => {
        setSource(undefined);
        if (burnID.current === undefined) setMode(compatible ? 'compatible' : 'original');
        setTarget(position);
      }}
      next={() => {
        if (source.info.next_episode) {
          setSource(undefined);
          setMode('original');
          setSubtitle(undefined);
          burnID.current = undefined;
          setTarget(undefined);
          setCurrentID(source.info.next_episode.id);
        }
      }}
    />
  );
}
