'use strict';

function seekPlayer(position) {
  if (!player.info) return;
  const target = Math.max(0, Math.min(position, player.info.item.duration - .2));
  if (player.info.plan.mode === 'direct' && player.ready) {
    setNativeCurrentTime(target);
    player.position = target;
  } else startPlayerStream(target);
}
function bindPlayerEvents() {
  const bind = (event, handler) => art.events.proxy(video, event, handler);
  bind('timeupdate', () => {
    if (player.ready && player.info && !player.loading) player.position = playbackPosition();
  });
  bind('playing', () => {
    clearTimeout(player.pauseTimer);
    $('player-status').textContent = '正在播放';
  });
  bind('waiting', () => { $('player-status').textContent = '正在缓冲…'; });
  bind('pause', () => {
    if (!player.info || player.loading || player.released) return;
    $('player-status').textContent = '已暂停';
    persistPlayback();
    clearTimeout(player.pauseTimer);
    if (player.info.plan.mode !== 'direct') player.pauseTimer = setTimeout(() => {
      if (!video || !video.paused || !player.info) return;
      player.position = playbackPosition();
      player.released = true;
      player.ready = false;
      clearPlayerSubtitles();
      video.removeAttribute('src');
      video.load();
    }, 15000);
  });
  bind('ended', () => {
    if (!player.info) return;
    const position = playbackPosition();
    const duration = player.info.item.duration;
    if (duration > 0 && duration - position < 5) {
      player.position = duration;
      player.ready = false;
      if (player.info.next_episode) playNextEpisode(true);
      else persistPlayback(true, true);
      $('player-status').textContent = '播放结束';
    } else {
      persistPlayback(true);
      player.position = position;
      player.ready = false;
      $('player-status').textContent = '播放已中断';
      showError('player-error', '视频流提前结束，可以重试兼容模式继续观看。');
    }
  });
  bind('error', () => {
    if (!player.info || !video.getAttribute('src')) return;
    $('player-status').textContent = '播放失败';
    showError('player-error', '视频加载失败，可重试兼容模式；也请确认 FFmpeg 可用且没有超过两个兼容播放任务。');
  });
}
$('player-retry').addEventListener('click', () => changePlayerSource({ mode: 'compatible' }));
document.addEventListener('keydown', event => {
  if (state.section === 'player' && event.code === 'Escape' && !document.fullscreenElement) {
    returnFromPlayer();
  }
});
