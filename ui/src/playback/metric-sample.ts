export type PlaybackMode = 'direct' | 'remux' | 'transcode';
export type PlaybackQuality = 'auto' | 'original' | '1080' | '720' | '480';

export interface PlaybackMetric {
  id: string;
  mode: PlaybackMode;
  quality: PlaybackQuality;
  stream_id?: string;
  burn_subtitle?: boolean;
  media_id?: number;
  compatibility_requested?: boolean;
  audio_index?: number;
  width: number;
  height: number;
  startup_ms?: number;
  stalls: number;
  stall_ms: number;
  played_ms: number;
  buffer_ms: number;
  dropped_frames?: number;
  total_frames?: number;
}

export interface PlaybackMetricSource {
  title?: string;
  path?: string;
  width?: number;
  height?: number;
  video_codec?: string;
  audio_codec?: string;
  container?: string;
  bitrate_bps?: number;
  size_bytes?: number;
}

export interface PlaybackMetricRow extends Omit<PlaybackMetric, 'stream_id'> {
  updated_at: number;
  transcode_speed?: number;
  source?: PlaybackMetricSource;
  reason?: string;
}

export interface MetricSelection {
  mediaId?: number;
  compatibilityRequested?: boolean;
  audioIndex?: number;
}

export function playbackMetricID() {
  // randomUUID() is limited to secure contexts; getRandomValues() works over HTTP.
  const bytes = crypto.getRandomValues(new Uint8Array(16));
  bytes[6] = (bytes[6] & 0x0f) | 0x40;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;
  const hex = Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
}

export function playbackQuality(value: string): PlaybackQuality {
  switch (value) {
    case 'original':
    case '1080':
    case '720':
    case '480':
      return value;
    default:
      return 'auto';
  }
}

export function bufferedAhead(ranges: TimeRanges, time: number) {
  for (let index = 0; index < ranges.length; index++) {
    if (ranges.start(index) <= time && time <= ranges.end(index)) return (ranges.end(index) - time) * 1000;
  }
  return 0;
}

const ms = (value: number) => (Number.isFinite(value) ? Math.max(0, Math.round(value)) : 0);

type Phase = 'playing' | 'stall' | 'idle';

export class MetricSample {
  private phase: Phase = 'idle';
  private phaseStart: number;
  private started = false;
  private autoplayBlocked = false;
  private seeking = false;
  private suspended = false;
  private played = 0;
  private stalled = 0;
  private data: PlaybackMetric;
  private loadTime: number;

  constructor(
    id: string,
    mode: PlaybackMode,
    quality: PlaybackQuality,
    loadTime: number,
    burnSubtitle = false,
    { mediaId, compatibilityRequested, audioIndex }: MetricSelection = {},
  ) {
    this.loadTime = loadTime;
    this.phaseStart = loadTime;
    this.data = {
      id,
      mode,
      quality,
      ...(mode !== 'direct' ? { stream_id: id } : {}),
      burn_subtitle: burnSubtitle,
      ...(mediaId != null && Number.isSafeInteger(mediaId) && mediaId > 0 ? { media_id: mediaId } : {}),
      ...(typeof compatibilityRequested === 'boolean'
        ? { compatibility_requested: compatibilityRequested }
        : {}),
      ...(audioIndex != null && Number.isSafeInteger(audioIndex) && audioIndex >= 0
        ? { audio_index: audioIndex }
        : {}),
      width: 0,
      height: 0,
      stalls: 0,
      stall_ms: 0,
      played_ms: 0,
      buffer_ms: 0,
    };
  }

  private transition(phase: Phase, now: number) {
    if (this.phase === phase) return;
    if (this.phase === 'playing') this.played += now - this.phaseStart;
    if (this.phase === 'stall') this.stalled += now - this.phaseStart;
    if (phase === 'stall') this.data.stalls++;
    this.phase = phase;
    this.phaseStart = now;
  }

  playing(now: number) {
    if (this.seeking || this.suspended) return;
    if (!this.started && !this.autoplayBlocked) this.data.startup_ms = ms(now - this.loadTime);
    this.started = true;
    this.transition('playing', now);
  }

  blockAutoplay() {
    this.autoplayBlocked = true;
  }

  waiting(now: number) {
    // Startup, seek buffering and repeated waiting for one stall are not new stalls.
    if (this.phase === 'playing' && !this.seeking && !this.suspended) this.transition('stall', now);
  }

  stop(now: number) {
    this.transition('idle', now);
  }

  beginSeek(now: number) {
    this.seeking = true;
    this.stop(now);
  }

  endSeek(now: number, canPlay: boolean) {
    this.seeking = false;
    if (this.started && canPlay) this.playing(now);
  }

  suspend(now: number) {
    this.suspended = true;
    this.stop(now);
  }

  resume(now: number, canPlay: boolean) {
    this.suspended = false;
    if (this.started && canPlay) this.playing(now);
  }

  readVideo(
    video: Pick<HTMLVideoElement, 'videoWidth' | 'videoHeight' | 'getVideoPlaybackQuality'>,
    buffer: number,
  ) {
    if (video.videoWidth && video.videoHeight) {
      this.data.width = video.videoWidth;
      this.data.height = video.videoHeight;
    }
    this.data.buffer_ms = ms(buffer);
    const frames = video.getVideoPlaybackQuality?.();
    if (frames && Number.isFinite(frames.droppedVideoFrames) && Number.isFinite(frames.totalVideoFrames)) {
      this.data.dropped_frames = Math.max(this.data.dropped_frames ?? 0, frames.droppedVideoFrames);
      this.data.total_frames = Math.max(this.data.total_frames ?? 0, frames.totalVideoFrames);
    }
  }

  snapshot(now: number): PlaybackMetric {
    const elapsed = now - this.phaseStart;
    return {
      ...this.data,
      played_ms: ms(this.played + (this.phase === 'playing' ? elapsed : 0)),
      stall_ms: ms(this.stalled + (this.phase === 'stall' ? elapsed : 0)),
    };
  }
}
