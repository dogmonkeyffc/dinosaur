import { useEffect, useState, type ReactNode } from 'react';
import { BackHandler, Platform, Pressable, StatusBar, Text, View, useWindowDimensions } from 'react-native';
import type { PlaybackInfo } from '@dinosaur/contracts';
import { VideoView, type VideoPlayer } from 'expo-video';
import { LinearGradient } from 'expo-linear-gradient';
import { SeekBar } from './SeekBar';
import { PhoneTransport, PlayerAction } from './PhoneControls';
import { PlayerMenu, type PlayerSetting } from './PlayerMenu';
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
  settings,
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
  settings: PlayerSetting[];
  children: ReactNode;
  error: string;
  fullscreen: { fullscreen: boolean; toggle: () => Promise<void> };
}) {
  const { fullscreen, toggle: rotate } = rotation;
  const [visible, setVisible] = useState(true);
  const [seeking, setSeeking] = useState(false);
  const [menu, setMenu] = useState<number>();
  const { width } = useWindowDimensions();
  useEffect(() => {
    if (!fullscreen || !playing || !visible || seeking || menu !== undefined) return;
    const timer = setTimeout(() => setVisible(false), 5000);
    return () => clearTimeout(timer);
  }, [fullscreen, playing, visible, seeking, menu]);
  useEffect(() => {
    const handler = BackHandler.addEventListener('hardwareBackPress', () => {
      if (menu !== undefined) setMenu(undefined);
      else if (fullscreen) void rotate();
      else close();
      return true;
    });
    return () => handler.remove();
  }, [fullscreen, rotate, close, menu]);
  const show = !fullscreen || visible || !playing || !!error || menu !== undefined;
  const title = info.item.series_title ? `${info.item.series_title} · ${info.item.title}` : info.item.title;
  return (
    <View
      style={[
        css.screen,
        !fullscreen && { paddingTop: Platform.OS === 'android' ? StatusBar.currentHeight : 52 },
      ]}
    >
      <StatusBar hidden={fullscreen} barStyle="light-content" />
      <View style={css.stage}>
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
        {show && (
          <>
            <LinearGradient colors={['#000b', 'transparent']} style={css.header}>
              <PlayerAction label="返回" icon="back" press={() => (fullscreen ? void rotate() : close())} />
              <Text numberOfLines={1} style={css.title}>
                {title}
              </Text>
            </LinearGradient>
            <LinearGradient colors={['transparent', '#000b', '#000e']} style={css.controls}>
              <SeekBar time={time} duration={info.item.duration} seek={seek} touching={setSeeking} />
              <View style={css.controlRow}>
                <PhoneTransport
                  playing={playing}
                  toggle={toggle}
                  seek={seek}
                  next={next}
                  wide={width > 720}
                />
                <View style={css.spacer} />
                {settings.map((setting, index) => (
                  <PlayerAction
                    key={setting.label}
                    label={setting.label}
                    icon={setting.icon}
                    disabled={setting.disabled}
                    press={() => setMenu((value) => (value === index ? undefined : index))}
                  />
                ))}
                <PlayerAction
                  label={fullscreen ? '退出全屏' : '全屏'}
                  icon={fullscreen ? 'collapse' : 'fullscreen'}
                  press={() => {
                    setVisible(true);
                    void rotate();
                  }}
                />
              </View>
            </LinearGradient>
          </>
        )}
        {menu !== undefined && (
          <>
            <Pressable
              accessibilityRole="button"
              accessibilityLabel="关闭播放设置"
              style={css.menuBackdrop}
              onPress={() => setMenu(undefined)}
            />
            <PlayerMenu setting={settings[menu]} close={() => setMenu(undefined)} />
          </>
        )}
        {!!error && (
          <Text accessibilityRole="alert" style={css.error}>
            {error}
          </Text>
        )}
      </View>
    </View>
  );
}
