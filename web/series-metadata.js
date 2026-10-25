'use strict';

function seriesCandidateCards(candidates, id) {
  return candidates.slice(0, 5).map(item => `<article class="metadata-candidate"><strong>${escapeHTML(item.title)} (${item.year || '年份未知'})</strong>
    <span>${escapeHTML(item.original_title)}</span><p>${escapeHTML(item.overview)}</p>
    <button class="button secondary" data-series-candidate="${item.id}" data-series-id="${id}" ${seriesScraping.active ? 'disabled' : ''}>选择此剧集 · ${item.id}</button></article>`).join('');
}
function seriesMetadataHTML(item) {
  const info = item.metadata;
  const matched = info?.tmdb_id > 0;
  const status = matched ? '已匹配 TMDB' : info?.status === 'needs_review' ? '需要确认匹配' : info?.status === 'failed' ? '刮削失败' : '尚未识别剧集资料';
  const description = `${info?.original_title ? `<p class="movie-original-title">${escapeHTML(info.original_title)}</p>` : ''}
    ${info?.overview ? `<p class="movie-overview">${escapeHTML(info.overview)}</p>` : ''}
    <p class="movie-metadata-status">${status} ${matched ? `<a href="https://www.themoviedb.org/tv/${info.tmdb_id}" target="_blank" rel="noopener noreferrer">TMDB · ${info.tmdb_id}</a>` : ''}</p>
    ${info?.message ? `<p class="field-help">${escapeHTML(scrapeMessage(info.message))}</p>` : ''}`;
  if (state.user?.role !== 'admin') return description;
  return `${description}<div class="metadata-detail-actions"><button class="text-button" data-series-scrape="${item.id}" data-refresh="${matched}" ${seriesScraping.active ? 'disabled' : ''}>${matched ? '刷新已匹配资料' : '自动识别'}</button></div>
    <details class="series-matching"><summary>匹配 / 更正剧集</summary>
      <form id="series-match-search" data-id="${item.id}"><label>剧名<input name="query" maxlength="128" placeholder="留空使用本地剧名"></label><label>首播年份<input name="year" type="number" min="0" max="2100" value="${item.year || ''}" placeholder="0 表示不限"></label><button class="button secondary">搜索候选</button></form>
      <form id="series-match-id" data-id="${item.id}"><label>TMDB 剧集 ID<input name="tmdb_id" type="number" min="1" max="2147483647" required placeholder="TMDB 剧集页面的数字 ID"></label><button class="button secondary">使用此 ID 匹配</button></form>
      <p id="series-match-error" class="form-error" role="alert" hidden></p>
      <div id="series-match-results" aria-live="polite">${seriesCandidateCards(info?.candidates || [], item.id)}</div>
    </details>`;
}
$('series-detail').addEventListener('click', event => {
  const scrape = event.target.closest('[data-series-scrape]');
  if (scrape) startSeriesScraping({ series_id: Number(scrape.dataset.seriesScrape), refresh: scrape.dataset.refresh === 'true' }, scrape);
  const candidate = event.target.closest('[data-series-candidate]');
  if (candidate) startSeriesScraping({ series_id: Number(candidate.dataset.seriesId), tmdb_id: Number(candidate.dataset.seriesCandidate) }, candidate);
});
$('series-detail').addEventListener('submit', async event => {
  const form = event.target;
  if (!['series-match-search', 'series-match-id'].includes(form.id)) return;
  event.preventDefault();
  const values = new FormData(form);
  const id = Number(form.dataset.id);
  if (form.id === 'series-match-id') {
    await startSeriesScraping({ series_id: id, tmdb_id: Number(values.get('tmdb_id')) }, event.submitter);
    return;
  }
  const userID = state.user?.id;
  const revision = series.detailRequest;
  const active = () => userID === state.user?.id && revision === series.detailRequest && state.section === 'series-detail';
  const query = new URLSearchParams();
  if (values.get('query').trim()) query.set('q', values.get('query').trim());
  if (values.get('year') !== '') query.set('year', values.get('year'));
  showError('series-match-error');
  await busy(event.submitter, '正在搜索…', async () => {
    try {
      const result = await api(`/api/admin/series/${id}/metadata/candidates?${query}`);
      if (!active()) return;
      $('series-match-results').innerHTML = result.candidates.length ? seriesCandidateCards(result.candidates, id) : '<p class="field-help">未找到候选，可调整剧名或年份，也可以直接填写 TMDB 剧集 ID。</p>';
    } catch (error) { if (active()) handleError(error, 'series-match-error'); }
  });
});
