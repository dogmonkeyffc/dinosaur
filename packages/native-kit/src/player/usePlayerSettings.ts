import { useState } from 'react';
import type { VideoPlayer } from 'expo-video';
import type { PlaybackInfo } from '@dinosaur/contracts';
import type { PlayerSetting } from './PlayerMenu';

export function usePlayerSettings({
  player,
  info,
  track,
  position,
  selectSubtitle,
  restart,
  quality,
  selectQuality,
}: {
  player: VideoPlayer;
  info: PlaybackInfo;
  track?: number;
  position: () => number;
  selectSubtitle: (track: number | undefined, position: number) => void;
  restart: (position: number, compatible: boolean) => void;
  quality: string;
  selectQuality: (quality: string, position: number) => void;
}): PlayerSetting[] {
  const [, refresh] = useState(0);
  const [rate, setRate] = useState(1);
  const audio = player.availableAudioTracks;
  return [
    {
      icon: 'audio',
      label: '音轨',
      disabled: !audio.length,
      options: audio.map((item, index) => ({
        label: item.label || item.language || `音轨 ${index + 1}`,
        selected: item.id === player.audioTrack?.id,
        press: () => {
          player.audioTrack = item;
          refresh((value) => value + 1);
        },
      })),
    },
    {
      icon: 'subtitle',
      label: '字幕',
      disabled: !info.subtitles.length,
      options: [
        {
          label: '关闭字幕',
          selected: track === undefined,
          press: () => selectSubtitle(undefined, position()),
        },
        ...info.subtitles.map((item) => ({
          label: item.label,
          selected: item.id === track,
          press: () => selectSubtitle(item.id, position()),
        })),
      ],
    },
    {
      icon: 'mode',
      label: '播放方式',
      options: [
        { label: '自动选择', selected: info.plan.mode === 'direct', press: () => restart(position(), false) },
        { label: '兼容模式', selected: info.plan.mode !== 'direct', press: () => restart(position(), true) },
      ],
    },
    {
      icon: 'quality',
      label: '分辨率',
      options: ['auto', 'original', '1080', '720', '480'].map((value) => ({
        label: value === 'auto' ? '自动' : value === 'original' ? '原分辨率' : `${value}p`,
        selected: value === quality,
        press: () => selectQuality(value, position()),
      })),
    },
    {
      icon: 'rate',
      label: '倍速',
      options: [0.5, 0.75, 1, 1.25, 1.5].map((value) => ({
        label: `${value}×`,
        selected: rate === value,
        press: () => {
          player.playbackRate = value;
          setRate(value);
        },
      })),
    },
  ];
}
