'use strict';

const player = {
  info: null, position: 0, offset: 0, ready: false, loading: false,
  revision: 0, released: false, pauseTimer: null, progressTimer: null,
  seekTimer: null, seekTarget: null, audio: null, mode: 'auto', subtitle: '', rate: 1,
};
let video = null;
const playerClock = seconds => {
  const value = Math.max(0, Math.floor(seconds || 0));
  const hours = Math.floor(value / 3600);
  const minutes = Math.floor(value % 3600 / 60);
  const rest = String(value % 60).padStart(2, '0');
  return hours ? `${hours}:${String(minutes).padStart(2, '0')}:${rest}` : `${minutes}:${rest}`;
};
function playbackPosition() {
  if (!player.ready || player.released) return player.position;
  const position = player.offset + (Number.isFinite(nativeCurrentTime()) ? nativeCurrentTime() : 0);
  return Math.min(position, player.info?.item.duration || position);
}
async function openPlayer(id, fromBeginning = false, replace = false) {
  if (state.section === 'player') closePlayer();
  const revision = ++player.revision;
  $('player-title').textContent = '正在准备播放…';
  showError('player-error');
  switchSection('player', id, { replace });
  try {
    const info = await api(`/api/playback/${id}`);
    if (revision !== player.revision || state.section !== 'player') return;
    player.info = info;
    $('player-title').textContent = playbackTitle(info.item);
    syncPlayerNavigation();
    document.title = `${playbackTitle(info.item)} · Dinosaur`;
    player.audio = info.plan.audio;
    player.mode = 'auto';
    player.subtitle = '';
    player.rate = 1;
    createPlayerUI();
    startPlayerStream(fromBeginning ? 0 : info.resume_position);
    clearInterval(player.progressTimer);
    player.progressTimer = setInterval(() => { if (!video.paused) persistPlayback(); }, 10000);
  } catch (error) { handleError(error, 'player-error'); }
}
function startPlayerStream(position) {
  if (!player.info) return;
  clearTimeout(player.pauseTimer);
  player.loading = true;
  player.ready = false;
  player.released = false;
  player.position = Math.max(0, Math.min(position, Math.max(0, player.info.item.duration - .2)));
  player.offset = player.info.plan.mode === 'direct' ? 0 : player.position;
  video.pause();
  clearPlayerSubtitles();
  video.removeAttribute('src');
  video.load();
  const url = new URL(player.info.stream_url, location.origin);
  if (player.info.plan.mode !== 'direct') url.searchParams.set('start', player.position.toFixed(3));
  url.searchParams.set('_', String(Date.now()));
  art.url = url.pathname + url.search;
  art.poster = player.info.item.poster_url || '';
  $('player-status').textContent = '正在加载视频…';
  video.onloadedmetadata = () => {
    if (!player.info) return;
    if (player.info.plan.mode === 'direct') setNativeCurrentTime(player.position);
    art.playbackRate = player.rate;
    syncPlaybackOptions();
    player.ready = true;
    player.loading = false;
    applyPlayerSubtitle();
    video.play().catch(() => { $('player-status').textContent = '点击播放开始观看'; });
  };
  const labels = { direct: '直接播放 · 原画', remux: '兼容封装 · 原画', transcode: '兼容转码 · 最高 720p' };
  if (art) art.notice.show = labels[player.info.plan.mode];
  showError('player-error');
}
async function changePlayerSource(selection = {}) {
  if (!player.info) return;
  const position = playbackPosition();
  const id = player.info.item.id;
  const revision = ++player.revision;
  const mode = selection.mode ?? player.mode;
  const audio = selection.audio ?? player.audio;
  const query = new URLSearchParams({ mode });
  if (audio !== null && audio !== undefined) query.set('audio', audio);
  try {
    const info = await api(`/api/playback/${id}?${query}`);
    if (revision !== player.revision || state.section !== 'player') return;
    player.info = info;
    player.audio = info.plan.audio;
    player.mode = mode;
    startPlayerStream(position);
  } catch (error) { handleError(error, 'player-error'); }
}
function closePlayer() {
  if (document.fullscreenElement) document.exitFullscreen().catch(() => {});
  const saved = persistPlayback(true);
  ++player.revision;
  clearInterval(player.progressTimer);
  clearTimeout(player.pauseTimer);
  clearTimeout(player.seekTimer);
  player.seekTarget = null;
  player.info = null;
  player.ready = false;
  if (video) {
    video.onloadedmetadata = null;
    video.pause();
    clearPlayerSubtitles();
    destroyPlayerUI();
  }
  Promise.resolve(saved).finally(() => {
    if (state.user && !$('movies-section').hidden) loadMovies();
    if (state.user && state.section === 'series') loadSeries();
  });
  return saved;
}

document.addEventListener('click', event => {
  const button = event.target.closest('[data-play-movie]');
  if (button) openPlayer(Number(button.dataset.playMovie), button.dataset.beginning === 'true');
});
