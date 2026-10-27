import { changeSource, selectSubtitle, type SourceSelection } from './source';
import { message } from '../lib/errors';
import { saveProgress } from '../lib/progress';
import type { PlaybackInfo } from '../lib/types';
import { escapeHTML, type ArtInstance } from './artplayer';
import { createTimelineVideo, currentTime, setCurrentTime } from './timeline';
import { applySubtitle, clearSubtitles } from './subtitles';
import { settingsControls } from './options';
import { syncOptions } from './option-menu';
import { bindEvents } from './events';

interface Callbacks {
  status: (message: string) => void;
  error: (message: string) => void;
  notify: (message: string) => void;
  back: () => void;
  next: (id: number) => void;
}
export class PlaybackEngine {
  art: ArtInstance;
  video: HTMLVideoElement;
  ready = false;
  loading = false;
  released = false;
  disposed = false;
  savedPosition = 0;
  offset = 0;
  seekTarget: number | null = null;
  audio: number;
  mode = 'auto';
  subtitle = '';
  rate = 1;
  subtitleRevision = 0;
  subtitleWindow = -1;
  pauseTimer?: number;
  private seekTimer?: number;
  private progressTimer: number;
  sourceRequest?: AbortController;

  constructor(
    public root: HTMLElement,
    public info: PlaybackInfo,
    public callbacks: Callbacks,
    beginning: boolean,
  ) {
    this.audio = info.plan.audio;
    const Artplayer = window.Artplayer;
    Artplayer.LOG_VERSION = false;
    Artplayer.CONTEXTMENU = false;
    Artplayer.RECONNECT_TIME_MAX = 0;
    Artplayer.SETTING_WIDTH = 200;
    Artplayer.PLAYBACK_RATE = [0.5, 0.75, 1, 1.25, 1.5];
    const title = info.item.series_id ? `${info.item.series_title} · ${info.item.title}` : info.item.title;
    this.art = new Artplayer({
      container: root,
      theme: '#c2df9e',
      lang: 'zh-cn',
      setting: false,
      playbackRate: false,
      fullscreen: true,
      pip: true,
      hotkey: true,
      mutex: false,
      backdrop: false,
      autoPlayback: false,
      playsInline: true,
      moreVideoAttr: { preload: 'metadata', playsInline: true },
      icons: {
        check:
          '<svg viewBox="0 0 24 24" width="24" height="24"><path d="M9 16.2L4.8 12l-1.4 1.4L9 19 21 7l-1.4-1.4z"/></svg>',
        config:
          '<svg viewBox="0 0 24 24" width="24" height="24"><path d="M3 6h18v2H3zm0 5h18v2H3zm0 5h18v2H3z"/></svg>',
        playbackRate:
          '<svg viewBox="0 0 24 24" width="24" height="24"><path d="M8 5v14l11-7zM4 5h2v14H4z"/></svg>',
      },
      layers: [
        {
          name: 'navigation',
          html: `<button class="player-back" type="button" aria-label="返回" title="返回"><svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="m10 5-7 7 7 7M3 12h18"/></svg></button><span>${escapeHTML(title)}</span>`,
          mounted: (element: HTMLElement) =>
            element.querySelector('button')!.addEventListener('click', callbacks.back),
        },
      ],
      proxy: () => createTimelineVideo(this),
      controls: [
        ...settingsControls(this),
        ...(info.next_episode
          ? [
              {
                name: 'next-episode',
                position: 'left',
                index: 13,
                html: '<svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><path d="M5 4v16l11-8zM18 4h2v16h-2z"/></svg>',
                tooltip: '下一集 · ' + info.next_episode.title,
                click: () => this.next(),
              },
            ]
          : []),
        {
          name: 'back-ten',
          position: 'left',
          index: 11,
          html: '−10秒',
          tooltip: '后退10秒',
          click: () => this.seek(this.position() - 10),
        },
        {
          name: 'forward-ten',
          position: 'left',
          index: 12,
          html: '+10秒',
          tooltip: '前进10秒',
          click: () => this.seek(this.position() + 10),
        },
      ],
    });
    this.video = this.art.video;
    bindEvents(this);
    this.start(beginning ? 0 : info.resume_position);
    this.progressTimer = window.setInterval(() => {
      if (!this.video.paused) this.persist();
    }, 10000);
  }

  position() {
    if (!this.ready || this.released || !this.video) return this.savedPosition;
    const position = this.offset + (Number.isFinite(currentTime(this.video)) ? currentTime(this.video) : 0);
    return Math.min(position, this.info.item.duration || position);
  }

  start(position: number) {
    if (this.disposed) return;
    clearTimeout(this.pauseTimer);
    this.loading = true;
    this.ready = false;
    this.released = false;
    this.savedPosition = Math.max(0, Math.min(position, Math.max(0, this.info.item.duration - 0.2)));
    this.offset = this.info.plan.mode === 'direct' ? 0 : this.savedPosition;
    this.video.pause();
    clearSubtitles(this);
    this.video.removeAttribute('src');
    this.video.load();
    const url = new URL(this.info.stream_url, location.origin);
    if (this.info.plan.mode !== 'direct') url.searchParams.set('start', this.savedPosition.toFixed(3));
    url.searchParams.set('_', String(Date.now()));
    this.art.url = url.pathname + url.search;
    this.art.poster = this.info.item.poster_url || '';
    this.callbacks.status('正在加载视频…');
    this.callbacks.error('');
    this.video.onloadedmetadata = () => {
      if (this.disposed) return;
      if (this.info.plan.mode === 'direct') setCurrentTime(this.video, this.savedPosition);
      this.art.playbackRate = this.rate;
      syncOptions(this);
      this.ready = true;
      this.loading = false;
      applySubtitle(this);
      this.video.play().catch(() => {
        if (!this.disposed) this.callbacks.status('点击播放开始观看');
      });
    };
    this.art.notice.show = {
      direct: '直接播放 · 原画',
      remux: '兼容封装 · 原画',
      transcode: '兼容转码 · 最高 720p',
    }[this.info.plan.mode];
  }

  seek(position: number) {
    if (!Number.isFinite(position) || this.disposed) return;
    const target = Math.max(0, Math.min(position, this.info.item.duration - 0.2));
    clearTimeout(this.seekTimer);
    if (this.info.plan.mode === 'direct' && this.ready) {
      setCurrentTime(this.video, target);
      this.savedPosition = target;
      return;
    }
    this.seekTarget = target;
    this.seekTimer = window.setTimeout(() => {
      this.seekTarget = null;
      this.start(target);
    }, 180);
  }

  changeSource(selection: SourceSelection = {}) {
    return changeSource(this, selection);
  }

  selectSubtitle(value: string) {
    return selectSubtitle(this, value);
  }

  persist(keepalive = false, ended = false) {
    if (!this.ready && !this.released && !ended) return Promise.resolve();
    return saveProgress(
      this.info.item.id,
      { position: ended ? this.info.item.duration : this.position(), ...(ended ? { watched: true } : {}) },
      keepalive,
    ).catch((error) => {
      if (!this.disposed) this.callbacks.status(message(error));
    });
  }

  async next(ended = false) {
    const next = this.info.next_episode;
    if (!next) return;
    await this.persist(true, ended);
    if (!this.disposed) this.callbacks.next(next.id);
  }

  destroy() {
    this.persist(true);
    this.disposed = true;
    this.sourceRequest?.abort();
    clearInterval(this.progressTimer);
    clearTimeout(this.pauseTimer);
    clearTimeout(this.seekTimer);
    this.video.onloadedmetadata = null;
    this.video.pause();
    clearSubtitles(this);
    if (document.fullscreenElement?.contains(this.root)) document.exitFullscreen().catch(() => {});
    this.art.destroy();
  }
}
