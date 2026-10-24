'use strict';

function switchSection(section, mediaID = null, { replace = false } = {}) {
  if (state.section === 'player' && section !== 'player') closePlayer();
  state.section = section;
  for (const name of ['movies', 'libraries', 'account', 'player']) {
    $(name + '-section').hidden = name !== section;
  }
  $('content').classList.toggle('is-playing', section === 'player');
  document.querySelectorAll('[data-section]').forEach(button => {
    const active = button.dataset.section === (section === 'player' ? 'movies' : section);
    button.classList.toggle('active', active);
    if (active) button.setAttribute('aria-current', 'page');
    else button.removeAttribute('aria-current');
  });
  const titles = { movies: '我的电影', libraries: '媒体库设置', account: '账户安全', player: '正在播放' };
  $('breadcrumb-current').textContent = section === 'movies' ? '我的电影' : '系统设置';
  document.title = `${titles[section]} · Dinosaur`;
  const path = section === 'player' ? `/watch/${mediaID}` : section === 'movies' ? '/movies' : '/settings';
  if (location.pathname !== path) history[replace ? 'replaceState' : 'pushState'](null, '', path);
  $('content-main').scrollTop = 0;
  if (section === 'movies' && state.user) loadMovies();
}

function restorePage(replace = false) {
  const watch = location.pathname.match(/^\/watch\/(\d+)$/);
  if (watch) openPlayer(Number(watch[1]), false, replace);
  else switchSection(location.pathname === '/movies' ? 'movies' : 'libraries', null, { replace });
}
window.addEventListener('popstate', () => {
  if (state.user) restorePage(true);
});
