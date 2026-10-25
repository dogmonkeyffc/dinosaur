'use strict';

const series = { page: 1, request: 0, timer: null, detailRequest: 0, id: null, detail: null, season: null };
function syncSeriesLibraries() {
  const selected = $('series-library-filter').value;
  const libraries = state.libraries.filter(item => item.type === 'tv');
  $('series-library-filter').innerHTML = '<option value="">全部剧集媒体库</option>' +
    libraries.map(item => `<option value="${item.id}">${escapeHTML(item.name)}</option>`).join('');
  if (libraries.some(item => String(item.id) === selected)) $('series-library-filter').value = selected;
  $('scan-series').disabled = !libraries.length;
}
function browseSeriesLibrary(id) {
  $('series-library-filter').value = String(id);
  $('series-search').value = '';
  series.page = 1;
  switchSection('series');
}
function seriesCard(item) {
  return `<button class="movie-card" data-series="${item.id}" aria-label="查看 ${escapeHTML(item.title)}">
    <span class="movie-artwork"><span class="movie-placeholder">${icon('tv')}<span>DINOSAUR SERIES</span></span>
      ${item.poster_url ? `<img class="movie-poster" src="${escapeHTML(item.poster_url)}" alt="${escapeHTML(item.title)} 海报" loading="lazy">` : ''}
      ${itemMetadataBadge(item)}<span class="movie-artwork-caption">${escapeHTML(item.library_name)}</span>
    </span><span class="movie-title">${escapeHTML(item.title)}</span>
    <span class="movie-meta">${item.year || '年份未知'} · ${item.season_count} 季 · ${item.episode_count} 个文件</span>
    <span class="movie-meta">已看 ${item.watched_count} / ${item.episode_count}</span></button>`;
}
async function loadSeries() {
  if (!state.user) return;
  const userID = state.user.id;
  const request = ++series.request;
  const query = new URLSearchParams({ page: series.page, page_size: 24 });
  if ($('series-library-filter').value) query.set('library_id', $('series-library-filter').value);
  if ($('series-search').value.trim()) query.set('q', $('series-search').value.trim());
  showError('series-error');
  $('series-grid').setAttribute('aria-busy', 'true');
  try {
    const page = await api('/api/series?' + query);
    if (userID !== state.user?.id || request !== series.request) return;
    if (page.total && (series.page - 1) * 24 >= page.total) { series.page = 1; return loadSeries(); }
    $('series-total').textContent = page.total;
    $('series-grid').innerHTML = page.items.map(seriesCard).join('');
    bindPosterFallbacks($('series-grid'));
    $('series-empty').hidden = page.items.length > 0;
    $('series-pagination').hidden = page.total <= 24;
    $('series-page').textContent = `${page.total ? series.page : 0} / ${Math.ceil(page.total / 24)}`;
    $('series-prev').disabled = series.page <= 1;
    $('series-next').disabled = series.page * 24 >= page.total;
    loadSeriesContinue();
    if (state.user.role === 'admin') loadSeriesScrapeJobs();
  } catch (error) {
    if (userID === state.user?.id && request === series.request) handleError(error, 'series-error');
  } finally {
    if (request === series.request) $('series-grid').removeAttribute('aria-busy');
  }
}
async function loadSeriesContinue() {
  const userID = state.user?.id;
  try {
    const { items } = await api('/api/playback/continue?type=tv');
    if (state.user?.id !== userID) return;
    $('series-continue').hidden = !items.length;
    $('series-continue-grid').innerHTML = items.map(item => movieCard({ ...item, title: `${item.series_title} · ${item.title}` }, true)).join('');
    bindPosterFallbacks($('series-continue-grid'));
  } catch (error) {
    if (state.user?.id === userID) { $('series-continue').hidden = true; if (error.status === 401) handleError(error); }
  }
}
$('series-grid').addEventListener('click', event => {
  const card = event.target.closest('[data-series]');
  if (card) openSeries(Number(card.dataset.series));
});
$('series-search').addEventListener('input', () => {
  clearTimeout(series.timer);
  series.timer = setTimeout(() => { series.page = 1; loadSeries(); }, 250);
});
$('series-library-filter').addEventListener('change', () => { series.page = 1; loadSeries(); });
$('series-prev').addEventListener('click', () => { series.page--; loadSeries(); });
$('series-next').addEventListener('click', () => { series.page++; loadSeries(); });
$('scan-series').addEventListener('click', () => busy($('scan-series'), '正在开始扫描…', () => {
  const selected = Number($('series-library-filter').value);
  return scanLibraries(state.libraries.filter(item => item.type === 'tv' && (!selected || item.id === selected)));
}));
