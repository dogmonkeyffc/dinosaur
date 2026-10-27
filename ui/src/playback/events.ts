import type { PlaybackEngine } from './engine';
import { clearSubtitles, refreshSubtitle } from './subtitles';

export function bindEvents(engine: PlaybackEngine) {
  const { video, art, callbacks } = engine;
  const bind = (event: string, handler: () => void) => art.events.proxy(video, event, handler);
  bind('timeupdate', () => {
    if (engine.ready && !engine.loading) engine.savedPosition = engine.position();
    refreshSubtitle(engine);
  });
  bind('playing', () => {
    clearTimeout(engine.pauseTimer);
    callbacks.status('正在播放');
  });
  bind('waiting', () => callbacks.status('正在缓冲…'));
  bind('pause', () => {
    if (engine.disposed || engine.loading || engine.released) return;
    callbacks.status('已暂停');
    engine.persist();
    clearTimeout(engine.pauseTimer);
    if (engine.info.plan.mode !== 'direct')
      engine.pauseTimer = window.setTimeout(() => {
        if (!video.paused || engine.disposed) return;
        engine.savedPosition = engine.position();
        engine.released = true;
        engine.ready = false;
        clearSubtitles(engine);
        video.removeAttribute('src');
        video.load();
      }, 15000);
  });
  bind('ended', () => {
    if (engine.disposed) return;
    const position = engine.position();
    const duration = engine.info.item.duration;
    if (duration > 0 && duration - position < 5) {
      engine.savedPosition = duration;
      if (engine.info.next_episode) engine.next(true);
      else engine.persist(true, true);
      engine.ready = false;
      callbacks.status('播放结束');
    } else {
      engine.persist(true);
      engine.savedPosition = position;
      engine.ready = false;
      callbacks.status('播放已中断');
      callbacks.error('视频流提前结束，可以重试兼容模式继续观看。');
    }
  });
  bind('error', () => {
    if (!engine.disposed && video.getAttribute('src'))
      callbacks.error('视频加载失败，可重试兼容模式；也请确认 FFmpeg 可用且没有超过两个兼容播放任务。');
  });
  art.events.proxy(window, 'pagehide', () => engine.persist(true));
  art.events.proxy(document, 'keydown', (event) => {
    if ((event as KeyboardEvent).key === 'Escape' && !event.defaultPrevented && !document.fullscreenElement)
      callbacks.back();
  });
}
