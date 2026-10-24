'use strict';

let playbackSave = Promise.resolve();
function persistPlayback(keepalive = false, ended = false) {
  if (!player.info || !state.user || (!player.ready && !player.released && !ended)) return;
  const mediaID = player.info.item.id;
  const userID = state.user.id;
  const body = { position: ended ? player.info.item.duration : playbackPosition() };
  if (ended) body.watched = true;
  // Serialize writes so a slower heartbeat cannot overwrite a later pause/seek.
  playbackSave = playbackSave.catch(() => {}).then(async () => {
    if (state.user?.id !== userID) return;
    const response = await fetch(`/api/playback/${mediaID}/progress`, {
      method: 'PUT', credentials: 'same-origin', keepalive,
      headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': state.csrf },
      body: JSON.stringify(body), signal: AbortSignal.timeout(10000),
    });
    if (state.user?.id !== userID) return;
    if (!response.ok) {
      if (response.status === 401) {
        const error = new Error('登录已过期，请重新登录。');
        error.status = 401;
        handleError(error);
      } else throw new Error('进度保存失败，稍后会自动重试。');
    }
  }).catch(error => {
    if (player.info) $('player-status').textContent = error.message;
  });
  return playbackSave;
}
window.addEventListener('pagehide', () => persistPlayback(true));
async function markMovieWatched(id, watched) {
  try {
    await api(`/api/playback/${id}/progress`, { method: 'PUT', body: { watched } });
    notify(watched ? '已标记为已看。' : '已重置为未看。');
    await openMovie(id);
    loadMovies();
  } catch (error) { handleError(error); }
}
document.addEventListener('click', event => {
  const button = event.target.closest('[data-watched-movie]');
  if (button) markMovieWatched(Number(button.dataset.watchedMovie), button.dataset.watched !== 'true');
});
