import { useEffect, useMemo, useRef, useState } from 'react';
import { AppState } from 'react-native';
import { useVideoPlayer } from 'expo-video';
import type { Client } from '@dinosaur/client';
import type { PlaybackInfo } from '@dinosaur/contracts';
import { ProgressQueue, sourcePosition } from '@dinosaur/playback-core';

export function usePlayback(
  client: Client,
  info: PlaybackInfo,
  start: number,
  failed: (position: number) => void,
  next: () => void,
) {
  const offset = info.plan.mode === 'direct' ? 0 : start;
  const path = info.stream_url + (info.stream_url.includes('?') ? '&' : '?') + 'start=' + offset;
  const source = useMemo(
    () => ({ ...client.resource(path), contentType: 'progressive' as const, useCaching: false }),
    [client, path],
  );
  const player = useVideoPlayer(source, (instance) => {
    instance.timeUpdateEventInterval = 1;
  });
  const position = useRef(start);
  const started = useRef(false);
  const ended = useRef(false);
  const [time, setTime] = useState(start);
  const [error, setError] = useState('');
  const queue = useMemo(
    () => new ProgressQueue((value, watched) => client.progress(info.item.id, value, watched)),
    [client, info.item.id],
  );
  useEffect(() => {
    let active = true;
    const report = (reason: unknown) => {
      if (active) setError(reason instanceof Error ? reason.message : '进度保存失败。');
    };
    const save = () => {
      if (started.current) queue.save(position.current).catch(report);
    };
    const listeners = [
      player.addListener('sourceLoad', ({ availableAudioTracks }) => {
        if (started.current) return;
        player.currentTime = info.plan.mode === 'direct' ? start : 0;
        if (info.plan.mode === 'direct') {
          const ordinal = info.audio_tracks.findIndex((track) => track.index === info.plan.audio);
          const selected = availableAudioTracks[ordinal];
          if (selected) player.audioTrack = selected;
        }
        player.subtitleTrack = null;
        started.current = true;
        player.play();
      }),
      player.addListener('timeUpdate', ({ currentTime }) => {
        if (!started.current) return;
        position.current = sourcePosition(currentTime, offset, info.item.duration);
        setTime(position.current);
      }),
      player.addListener('playingChange', ({ isPlaying }) => {
        if (!isPlaying) save();
      }),
      player.addListener('statusChange', ({ status }) => {
        if (status === 'error') failed(position.current);
      }),
      player.addListener('playToEnd', () => {
        if (ended.current) return;
        if (info.item.duration > 0 && position.current < info.item.duration - 5) {
          failed(position.current);
          return;
        }
        ended.current = true;
        position.current = info.item.duration || position.current;
        queue
          .save(position.current, true)
          .then(() => {
            if (active) next();
          })
          .catch(report);
      }),
    ];
    const app = AppState.addEventListener('change', (state) => {
      if (state !== 'active') {
        player.pause();
        save();
      }
    });
    const interval = setInterval(() => {
      if (player.playing && !ended.current) save();
    }, 10000);
    return () => {
      active = false;
      clearInterval(interval);
      app.remove();
      listeners.forEach((listener) => listener.remove());
      save();
    };
  }, [player, queue, info.item.id]);
  async function flush() {
    if (started.current) await queue.save(position.current);
    else await queue.flush();
  }
  return { player, time, error, flush, position };
}
