'use strict';

function playbackTitle(item) {
  return item.series_id ? `${item.series_title} · ${item.title}` : item.title;
}
function returnFromPlayer() {
  const id = player.info?.item.series_id;
  if (id) openSeries(id, { replace: true });
  else switchSection('movies', null, { replace: true });
}
function syncPlayerNavigation() {
  const selected = player.info?.item.series_id ? 'series' : 'movies';
  document.querySelectorAll('[data-section]').forEach(button => {
    const active = button.dataset.section === selected;
    button.classList.toggle('active', active);
    if (active) button.setAttribute('aria-current', 'page');
    else button.removeAttribute('aria-current');
  });
}
async function playNextEpisode(ended = false) {
  const next = player.info?.next_episode;
  const revision = player.revision;
  if (!next) return;
  await persistPlayback(true, ended);
  if (revision !== player.revision || state.section !== 'player') return;
  await openPlayer(next.id, true, true);
}
function nextEpisodeControl() {
  if (!player.info?.next_episode) return [];
  return [{ name: 'next-episode', position: 'left', index: 13, html: '<svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M5 4v16l11-8zM18 4h2v16h-2z"/></svg>', tooltip: '下一集 · ' + player.info.next_episode.title, click: () => playNextEpisode() }];
}
