import { useEffect, useState } from 'react';
import { BackHandler, Text, View, useWindowDimensions } from 'react-native';
import { VideoView } from 'expo-video';
import type { Client } from '@dinosaur/client';
import type { PlaybackInfo } from '@dinosaur/contracts';
import { clock } from '@dinosaur/playback-core';
import { Button, ErrorText } from '../components';
import { styles } from '../theme';
import { usePlayback } from './usePlayback';
import { Subtitles } from './Subtitles';
import { PhoneTransport, PlayerAction, phonePlayerStyles as phone } from './PhoneControls';

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
  const { width, height } = useWindowDimensions();
  const portrait = height >= width;
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
  const toggle = () => {
    if (playback.player.playing) playback.player.pause();
    else playback.player.play();
  };
  async function nextEpisode() {
    try {
      await playback.flush();
      next();
    } catch {
      setError('进度保存失败。');
    }
  }
  const useCompatible = () => restart(playback.position.current, true);
  return (
    <View style={[
      styles.screen,
      tv ? { padding: 28, gap: 14 } : phone.screen,
      !tv && !portrait && { paddingTop: 16, gap: 4 },
    ]}>
      <View style={styles.row}>
        {tv ? <Button onPress={exit}>返回</Button> : <PlayerAction label="返回" text="←" press={exit} />}
        <Text numberOfLines={tv ? undefined : 1} style={[styles.subtitle, !tv && phone.title]}>
          {info.item.title}
        </Text>
      </View>
      <View style={tv
        ? { flex: 1, minHeight: 200, backgroundColor: '#000' }
        : [phone.video, portrait ? phone.portraitVideo : phone.landscapeVideo]}>
        <VideoView
          player={playback.player}
          nativeControls
          fullscreenOptions={{ enable: true, orientation: 'landscape' }}
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
      {tv ? (
        <View style={styles.row}>
          <Button preferred onPress={toggle}>播放 / 暂停</Button>
          <Button onPress={() => seek(-10)}>后退 10 秒</Button>
          <Button onPress={() => seek(30)}>前进 30 秒</Button>
          <Button onPress={audioTrack}>{audio}</Button>
          {info.subtitles.length > 0 && <Button onPress={subtitle}>字幕：{selected?.label || '关闭'}</Button>}
          {!compatible && <Button onPress={useCompatible}>兼容播放</Button>}
          {info.next_episode && <Button onPress={nextEpisode}>下一集</Button>}
        </View>
      ) : (
        <>
          <PhoneTransport playing={playback.playing} toggle={toggle} seek={seek} />
          <View style={phone.options}>
            <PlayerAction label={`切换音轨，${audio}`} text="音轨" press={audioTrack} />
            {info.subtitles.length > 0 && (
              <PlayerAction label={`切换字幕，${selected?.label || '关闭'}`} text={selected ? `字幕 ${info.subtitles.indexOf(selected) + 1}/${info.subtitles.length}` : '字幕 · 关'} press={subtitle} />
            )}
            {!compatible && <PlayerAction label="兼容播放" text="兼容播放" press={useCompatible} />}
            {info.next_episode && <PlayerAction label="下一集" text="下一集" press={nextEpisode} />}
          </View>
        </>
      )}
      <ErrorText error={error || playback.error} />
    </View>
  );
}
