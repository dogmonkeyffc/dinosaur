import { useEffect, useState, type ReactNode } from 'react';
import { BackHandler, Platform, Pressable, StatusBar, Text, View } from 'react-native';
import type { PlaybackInfo } from '@dinosaur/contracts';
import type { VideoPlayer } from 'expo-video';
import { VideoView } from 'expo-video';
import { SeekBar } from './SeekBar';
import { PhoneTransport, PlayerAction } from './PhoneControls';
import { phoneStyles as css } from './phoneStyles';

export function PhonePlayer({
  info,
  player,
  time,
  playing,
  close,
  seek,
  toggle,
  next,
  audio,
  subtitle,
  compatible,
  children,
  error,
  fullscreen: rotation,
}: {
  info: PlaybackInfo;
  player: VideoPlayer;
  time: number;
  playing: boolean;
  close: () => void;
  seek: (delta: number) => void;
  toggle: () => void;
  next?: () => void;
  audio: { label: string; press: () => void };
  subtitle?: { label: string; press: () => void };
  compatible?: () => void;
  children: ReactNode;
  error: string;
  fullscreen: { fullscreen: boolean; toggle: () => Promise<void> };
}) {
  const { fullscreen, toggle: rotate } = rotation;
  const [visible, setVisible] = useState(true);
  const [seeking, setSeeking] = useState(false);
  useEffect(() => {
    if (!fullscreen || !playing || !visible || seeking) return;
    const timer = setTimeout(() => setVisible(false), 5000);
    return () => clearTimeout(timer);
  }, [fullscreen, playing, visible, seeking]);
  useEffect(() => {
    const handler = BackHandler.addEventListener('hardwareBackPress', () => {
      if (fullscreen) void rotate();
      else close();
      return true;
    });
    return () => handler.remove();
  }, [fullscreen, rotate, close]);
  const show = !fullscreen || visible || !playing || !!error;
  return (
    <View
      style={[
        css.screen,
        fullscreen
          ? css.fullscreen
          : { paddingTop: Platform.OS === 'android' ? StatusBar.currentHeight : 52 },
      ]}
    >
      <StatusBar hidden={fullscreen} barStyle="light-content" />
      <View style={[css.header, fullscreen && css.floatingHeader, fullscreen && !show && css.hidden]}>
        <PlayerAction label="返回" text="←" press={() => (fullscreen ? void rotate() : close())} />
        <Text numberOfLines={1} style={css.title}>
          {info.item.title}
        </Text>
      </View>
      <View style={[css.frame, fullscreen ? css.fullFrame : css.portraitFrame]}>
        <VideoView
          player={player}
          nativeControls={false}
          contentFit="contain"
          surfaceType="textureView"
          style={css.video}
        />
        {children}
        <Pressable
          accessibilityRole="button"
          accessibilityLabel="显示播放控制"
          style={css.tapArea}
          onPress={() => setVisible((value) => !value)}
        />
      </View>
      <View style={[css.controls, fullscreen && css.floatingControls, fullscreen && !show && css.hidden]}>
        <PhoneTransport playing={playing} toggle={toggle} seek={seek} next={next} />
        <SeekBar time={time} duration={info.item.duration} seek={seek} touching={setSeeking} />
        <View style={css.options}>
          {info.audio_tracks.length > 1 && (
            <PlayerAction label={`切换音轨，${audio.label}`} text={audio.label} press={audio.press} />
          )}
          {subtitle && (
            <PlayerAction
              label={`切换字幕，${subtitle.label}`}
              text={subtitle.label}
              press={subtitle.press}
            />
          )}
          {compatible && <PlayerAction label="兼容播放" text="兼容播放" press={compatible} />}
          <PlayerAction
            label={fullscreen ? '退出全屏' : '全屏'}
            text={fullscreen ? '退出全屏 ⛶' : '全屏 ⛶'}
            press={() => {
              setVisible(true);
              void rotate();
            }}
          />
        </View>
        {!!error && (
          <Text accessibilityRole="alert" style={css.error}>
            {error}
          </Text>
        )}
      </View>
    </View>
  );
}
