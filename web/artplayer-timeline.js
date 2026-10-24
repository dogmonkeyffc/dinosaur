'use strict';

const nativeMedia = {
  time: Object.getOwnPropertyDescriptor(HTMLMediaElement.prototype, 'currentTime'),
  duration: Object.getOwnPropertyDescriptor(HTMLMediaElement.prototype, 'duration'),
  buffered: Object.getOwnPropertyDescriptor(HTMLMediaElement.prototype, 'buffered'),
  play: HTMLMediaElement.prototype.play,
};
function nativeCurrentTime(element = video) {
  return element ? nativeMedia.time.get.call(element) : 0;
}
function setNativeCurrentTime(value) {
  if (video) nativeMedia.time.set.call(video, value);
}

// Expose a source-wide timeline to Artplayer while keeping the browser's media
// clock relative to the currently generated FFmpeg segment.
function createTimelineVideo(callbacks) {
  const element = document.createElement('video');
  element.id = 'player-video';
  element.setAttribute('aria-label', '电影视频');
  Object.defineProperties(element, {
    duration: { get: callbacks.duration },
    currentTime: { get: callbacks.position, set: callbacks.seek },
    buffered: {
      get() {
        const ranges = nativeMedia.buffered.get.call(element);
        const offset = callbacks.offset();
        return {
          length: ranges.length,
          start: index => ranges.start(index) + offset,
          end: index => ranges.end(index) + offset,
        };
      },
    },
  });
  element.play = () => {
    if (callbacks.suspended() || element.error || element.ended) {
      callbacks.resume(element.ended ? 0 : callbacks.position());
      return Promise.resolve();
    }
    return nativeMedia.play.call(element);
  };
  return element;
}
function queuePlayerSeek(position) {
  if (!player.info || !Number.isFinite(position)) return;
  const target = Math.max(0, Math.min(position, player.info.item.duration - .2));
  clearTimeout(player.seekTimer);
  if (player.info.plan.mode === 'direct' && player.ready) { seekPlayer(target); return; }
  player.seekTarget = target;
  player.seekTimer = setTimeout(() => {
    player.seekTarget = null;
    if (state.section === 'player') seekPlayer(target);
  }, 180);
}
