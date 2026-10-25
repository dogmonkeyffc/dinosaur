'use strict';

function switchSection(section, mediaID = null, { replace = false } = {}) {
  const previousSection = state.section;
  if (state.section === 'detail' && section !== 'detail') movieDetailRequest++;
  if (['metadata', 'scraping'].includes(section) && state.user?.role !== 'admin') section = 'libraries';
  if (state.section === 'scraping' && section !== 'scraping') stopScrapeAdmin();
  if (state.section === 'metadata' && section !== 'metadata') resetMetadataSettings();
  const playbackSaved = state.section === 'player' && section !== 'player' ? closePlayer() : undefined;
  state.section = section;
  for (const name of ['movies', 'libraries', 'metadata', 'scraping', 'account', 'detail', 'player']) {
    $(name + '-section').hidden = name !== section;
  }
  $('content').classList.toggle('is-playing', section === 'player');
  $('content').classList.toggle('is-detail', section === 'detail');
  document.querySelectorAll('[data-section]').forEach(button => {
    const active = button.dataset.section === (['player', 'detail'].includes(section) ? 'movies' : section);
    button.classList.toggle('active', active);
    if (active) button.setAttribute('aria-current', 'page');
    else button.removeAttribute('aria-current');
  });
  const titles = { movies: '我的电影', libraries: '媒体库设置', account: '账户安全', metadata: '元数据设置', scraping: '刮削管理', player: '正在播放', detail: '电影详情' };
  document.title = `${titles[section]} · Dinosaur`;
  const path = section === 'detail' ? `/movies/${mediaID}` : section === 'player' ? `/watch/${mediaID}` : section === 'movies' ? '/movies' : section === 'metadata' ? '/settings/metadata' : section === 'scraping' ? '/settings/scraping' : '/settings';
  if (location.pathname !== path) {
    const routeState = section === 'detail' ? { detailOrigin: movieDetailOrigin, detailBack: Boolean(previousSection) && !replace } : null;
    history[replace ? 'replaceState' : 'pushState'](routeState, '', path);
  }
  $('content-main').scrollTop = 0;
  if (section === 'movies' && state.user) loadMovies();
  if (section === 'metadata' && state.user) loadMetadataSettings();
  if (section === 'scraping' && state.user) { syncScrapeAdminLibraries(); loadScrapeAdmin(); loadScrapeJobs(); }
  return playbackSaved;
}

function restorePage(replace = false) {
  const detail = location.pathname.match(/^\/movies\/(\d+)$/);
  if (detail) { openMovie(Number(detail[1]), { replace }); return; }
  const watch = location.pathname.match(/^\/watch\/(\d+)$/);
  if (watch) openPlayer(Number(watch[1]), false, replace);
  else switchSection(location.pathname === '/movies' ? 'movies' : location.pathname === '/settings/metadata' ? 'metadata' : location.pathname === '/settings/scraping' ? 'scraping' : 'libraries', null, { replace });
}
window.addEventListener('popstate', () => {
  if (state.user) restorePage(true);
});
