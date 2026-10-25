'use strict';

let art = null;
function createPlayerUI() {
  Artplayer.LOG_VERSION = false;
  Artplayer.CONTEXTMENU = false;
  Artplayer.RECONNECT_TIME_MAX = 0;
  Artplayer.SETTING_WIDTH = 200;
  Artplayer.PLAYBACK_RATE = [.5, .75, 1, 1.25, 1.5];
  art = new Artplayer({
    container: '#player-art',
    theme: '#c2df9e',
    lang: 'zh-cn',
    // These upstream icons use inline styles; supply CSP-compatible SVGs.
    icons: {
      check: '<svg viewBox="0 0 24 24" width="24" height="24"><path d="M9 16.2L4.8 12l-1.4 1.4L9 19 21 7l-1.4-1.4z"/></svg>',
      config: '<svg viewBox="0 0 24 24" width="24" height="24"><path d="M3 6h18v2H3zm0 5h18v2H3zm0 5h18v2H3z"/></svg>',
      playbackRate: '<svg viewBox="0 0 24 24" width="24" height="24"><path d="M8 5v14l11-7zM4 5h2v14H4z"/></svg>',
    },
    setting: false,
    layers: [{
      name: 'navigation',
      html: `<button class="player-back" type="button" aria-label="返回电影列表" title="返回电影列表"><svg width="22" height="22" viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="m10 5-7 7 7 7M3 12h18"/></svg></button><span>${escapeHTML(player.info.item.title)}</span>`,
      mounted: element => element.querySelector('button').addEventListener('click', () => switchSection('movies', null, { replace: true })),
    }],
    playbackRate: false,
    fullscreen: true,
    pip: true,
    hotkey: true,
    mutex: false,
    backdrop: false,
    autoPlayback: false,
    playsInline: true,
    moreVideoAttr: { preload: 'metadata', playsInline: true },
    proxy: () => createTimelineVideo({
      duration: () => player.info?.item.duration || 0,
      position: () => player.seekTarget ?? playbackPosition(),
      offset: () => player.offset,
      seek: queuePlayerSeek,
      suspended: () => player.released,
      resume: position => startPlayerStream(position),
    }),
    controls: [
      ...playerSettingsControls(),
      { name: 'back-ten', position: 'left', index: 11, html: '−10秒', tooltip: '后退10秒', click: () => queuePlayerSeek(playbackPosition() - 10) },
      { name: 'forward-ten', position: 'left', index: 12, html: '+10秒', tooltip: '前进10秒', click: () => queuePlayerSeek(playbackPosition() + 10) },
    ],
  });
  video = art.video;
  bindPlayerEvents();
}
function destroyPlayerUI() {
  if (!art) return;
  art.destroy();
  art = null;
  video = null;
}
