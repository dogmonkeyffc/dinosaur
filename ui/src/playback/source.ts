import { api } from '../lib/api';
import { message } from '../lib/errors';
import type { PlaybackInfo } from '../lib/types';
import type { PlaybackEngine } from './engine';
import { applySubtitle } from './subtitles';

export interface SourceSelection {
  mode?: string;
  audio?: number;
  subtitle?: string;
}

export async function changeSource(engine: PlaybackEngine, selection: SourceSelection) {
  const position = engine.position();
  const mode = selection.mode ?? engine.mode;
  const audio = selection.audio ?? engine.audio;
  const subtitle = selection.subtitle ?? engine.subtitle;
  const track = engine.info.subtitles.find((item) => String(item.id) === subtitle);
  engine.sourceRequest?.abort();
  const request = new AbortController();
  engine.sourceRequest = request;
  const query = new URLSearchParams({ mode });
  if (audio >= 0) query.set('audio', String(audio));
  if (track?.delivery === 'burn') query.set('subtitle', subtitle);
  try {
    const info = await api<PlaybackInfo>(`/api/playback/${engine.info.item.id}?${query}`, {
      signal: request.signal,
    });
    if (engine.disposed || request.signal.aborted) return;
    if (track?.delivery === 'burn' && engine.info.plan.subtitle === undefined)
      engine.callbacks.notify('图片字幕将通过兼容转码显示。');
    engine.info = info;
    engine.audio = info.plan.audio;
    engine.mode = mode;
    engine.subtitle = subtitle;
    engine.start(position);
  } catch (error) {
    if (!engine.disposed && !request.signal.aborted) engine.callbacks.error(message(error));
  } finally {
    if (engine.sourceRequest === request) engine.sourceRequest = undefined;
  }
}

export async function selectSubtitle(engine: PlaybackEngine, value: string) {
  const track = engine.info.subtitles.find((item) => String(item.id) === value);
  if (value && !track) return;
  if (track?.delivery === 'burn' || engine.info.plan.subtitle !== undefined || engine.sourceRequest) {
    await changeSource(engine, { subtitle: value });
  } else {
    engine.subtitle = value;
    applySubtitle(engine);
  }
}
