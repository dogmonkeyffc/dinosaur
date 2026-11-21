import { api, currentSession } from '../lib/api';
import type { ArtInstance } from './artplayer';
import {
  bufferedAhead,
  MetricSample,
  playbackMetricID,
  playbackQuality,
  type PlaybackMode,
} from './metric-sample';
import { buffered, currentTime } from './timeline';

interface SourceMetrics {
  sample: MetricSample;
  url: string;
  epoch: number;
  pending?: { key: string; keepalive: boolean };
  saved?: string;
}

export class PlaybackMetrics {
  private source?: SourceMetrics;
  private timer: number;

  constructor(
    art: ArtInstance,
    private video: HTMLVideoElement,
  ) {
    const bind = (event: string, handler: (sample: MetricSample, now: number) => void) => {
      art.events.proxy(video, event, () => {
        if (this.source && this.matches(this.source)) handler(this.source.sample, performance.now());
      });
    };
    bind('playing', (sample, now) => {
      if (this.canPlay()) sample.playing(now);
    });
    bind('waiting', (sample, now) => {
      if (!video.paused && !video.seeking && !video.ended) sample.waiting(now);
    });
    bind('seeking', (sample, now) => sample.beginSeek(now));
    bind('seeked', (sample, now) => sample.endSeek(now, this.canPlay()));
    for (const event of ['pause', 'ended', 'error']) bind(event, (sample, now) => sample.stop(now));
    for (const event of ['ended', 'error']) bind(event, () => this.report(true));
    art.events.proxy(window, 'pagehide', () => {
      this.source?.sample.suspend(performance.now());
      this.report(true);
    });
    art.events.proxy(window, 'pageshow', () => {
      this.source?.sample.resume(performance.now(), this.canPlay());
    });
    this.timer = window.setInterval(() => this.report(), 15000);
  }

  private canPlay() {
    const { video } = this;
    return !video.paused && !video.seeking && !video.ended && video.readyState >= video.HAVE_FUTURE_DATA;
  }

  private matches(source: SourceMetrics) {
    return this.video.currentSrc === source.url;
  }

  start(url: URL, mode: PlaybackMode, quality: string, burnSubtitle: boolean) {
    const id = playbackMetricID();
    if (mode !== 'direct') url.searchParams.set('stream_id', id);
    const source = {
      sample: new MetricSample(id, mode, playbackQuality(quality), performance.now(), burnSubtitle),
      url: new URL(url.pathname + url.search, location.origin).href,
      epoch: currentSession().epoch,
    };
    this.source = source;
    return (reason: unknown) => {
      if (this.source === source && reason instanceof DOMException && reason.name === 'NotAllowedError')
        source.sample.blockAutoplay();
    };
  }

  nativeSeek() {
    if (this.source && this.video.seeking) this.source.sample.beginSeek(performance.now());
  }

  private report(keepalive = false) {
    const source = this.source;
    const identity = currentSession();
    if (!source || !identity.session || identity.epoch !== source.epoch || (!keepalive && source.pending))
      return;
    const { video } = this;
    if (this.matches(source) && video.readyState >= video.HAVE_METADATA)
      source.sample.readVideo(video, bufferedAhead(buffered(video), currentTime(video)));
    const body = source.sample.snapshot(performance.now());
    const key = JSON.stringify(body);
    if (
      (!source.pending && key === source.saved) ||
      (source.pending?.keepalive && key === source.pending.key)
    )
      return;
    const request = { key, keepalive };
    source.pending = request;
    void api('/api/playback/metrics', { method: 'POST', body, keepalive })
      .then(() => {
        if (source.pending === request) source.saved = key;
      })
      .catch(() => {})
      .finally(() => {
        if (source.pending === request) source.pending = undefined;
      });
  }

  finish() {
    this.source?.sample.stop(performance.now());
    this.report(true);
    this.source = undefined;
  }

  destroy() {
    this.finish();
    clearInterval(this.timer);
  }
}
