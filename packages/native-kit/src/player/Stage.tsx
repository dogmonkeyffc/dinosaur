import { useEffect, useState } from 'react';
import { BackHandler, Text, View } from 'react-native';
import { VideoView } from 'expo-video';
import type { Client } from '@dinosaur/client';
import type { PlaybackInfo } from '@dinosaur/contracts';
import { clock } from '@dinosaur/playback-core';
import { Button, ErrorText } from '../components';
import { styles } from '../theme';
import { usePlayback } from './usePlayback';
import { Subtitles } from './Subtitles';

export function Stage({
  client,
  info,
  start,
  close,
  restart,
  next,
  tv,
  track,
  selectSubtitle,
}: {
  client: Client;
  info: PlaybackInfo;
  start: number;
  close: () => void;
  restart: (position: number, compatible: boolean) => void;
  next: () => void;
  tv: boolean;
  track?: number;
  selectSubtitle: (track: number | undefined, position: number) => void;
}) {
  const [error, setError] = useState('');
  const [audio, setAudio] = useState('音轨');
  const compatible = info.plan.mode !== 'direct';
  const playback = usePlayback(
    client,
    info,
    start,
    (position) => {
      if (!compatible) restart(position, true);
      else setError('无法播放此视频，请返回后重试。');
    },
    next,
  );
  async function exit() {
    try {
      playback.player.pause();
      await playback.flush();
    } catch {
      // A revoked or offline session must not trap the user in the player.
    } finally {
      close();
    }
  }
  useEffect(() => {
    const listener = BackHandler.addEventListener('hardwareBackPress', () => {
      exit();
      return true;
    });
    return () => listener.remove();
  }, [playback.player]);
  function seek(delta: number) {
    const target = Math.max(
      0,
      Math.min(Math.max(0, info.item.duration - 1), playback.position.current + delta),
    );
    if (compatible) restart(target, true);
    else playback.player.currentTime = target;
  }
  function audioTrack() {
    const tracks = playback.player.availableAudioTracks;
    if (!tracks.length) return;
    const index = tracks.findIndex((value) => value.id === playback.player.audioTrack?.id);
    const selected = tracks[(index + 1) % tracks.length];
    playback.player.audioTrack = selected;
    setAudio(selected.label || selected.language || '音轨');
  }
  function subtitle() {
    const available = info.subtitles;
    const index = available.findIndex((value) => value.id === track);
    selectSubtitle(
      index + 1 >= available.length ? undefined : available[index + 1].id,
      playback.position.current,
    );
  }
  const selected = info.subtitles.find((value) => value.id === track);
  return (
    <View style={[styles.screen, { padding: tv ? 28 : 16, paddingTop: tv ? 28 : 48, gap: 14 }]}>
      <View style={styles.row}>
        <Button onPress={exit}>返回</Button>
        <Text style={styles.subtitle}>{info.item.title}</Text>
      </View>
      <View style={{ flex: 1, minHeight: 200, backgroundColor: '#000' }}>
        <VideoView
          player={playback.player}
          nativeControls
          focusable
          style={{ flex: 1 }}
          contentFit="contain"
          surfaceType="textureView"
        />
        <Subtitles
          client={client}
          id={info.item.id}
          track={selected?.delivery === 'burn' ? undefined : track}
          embedded={selected?.kind === 'embedded'}
          time={playback.time}
          failed={setError}
        />
      </View>
      <Text style={styles.muted}>
        {clock(playback.time)} / {clock(info.item.duration)} · {compatible ? '兼容播放' : '原片播放'}
      </Text>
      <View style={styles.row}>
        <Button
          preferred
          onPress={() => {
            if (playback.player.playing) playback.player.pause();
            else playback.player.play();
          }}
        >
          播放 / 暂停
        </Button>
        <Button onPress={() => seek(-10)}>后退 10 秒</Button>
        <Button onPress={() => seek(30)}>前进 30 秒</Button>
        <Button onPress={audioTrack}>{audio}</Button>
        {info.subtitles.length > 0 && (
          <Button onPress={subtitle}>
            字幕：{info.subtitles.find((value) => value.id === track)?.label || '关闭'}
          </Button>
        )}
        {!compatible && <Button onPress={() => restart(playback.position.current, true)}>兼容播放</Button>}
        {info.next_episode && (
          <Button
            onPress={async () => {
              try {
                await playback.flush();
                next();
              } catch {
                setError('进度保存失败。');
              }
            }}
          >
            下一集
          </Button>
        )}
      </View>
      <ErrorText error={error || playback.error} />
    </View>
  );
}
