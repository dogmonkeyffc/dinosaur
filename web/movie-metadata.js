'use strict';

function movieMetadataHTML(item) {
  const info = item.metadata;
  const matched = info?.tmdb_id > 0;
  const status = matched ? '已匹配 TMDB' : info?.status === 'needs_review' ? '需要确认匹配' : info?.status === 'failed' ? '刮削失败' : '尚未识别电影资料';
  const description = info?.overview ? `<p class="movie-overview">${escapeHTML(info.overview)}</p>` : '';
  const original = info?.original_title && info.original_title !== item.title ? `<p class="movie-original-title">${escapeHTML(info.original_title)}</p>` : '';
  const source = matched ? `<a href="https://www.themoviedb.org/movie/${info.tmdb_id}" target="_blank" rel="noopener noreferrer">TMDB · ${info.tmdb_id}</a>` : '';
  const statusHTML = `<p class="movie-metadata-status">${status} ${source}</p>${info?.message ? `<p class="field-help">${escapeHTML(scrapeMessage(info.message))}</p>` : ''}`;
  if (state.user?.role !== 'admin') return original + description + statusHTML;
  return `${original}${description}${statusHTML}
    <div class="metadata-detail-actions"><button class="text-button" data-scrape-media="${item.id}" data-refresh="${matched}">${matched ? '刷新已匹配资料' : '自动识别'}</button></div>
    <details class="movie-matching"><summary>匹配 / 更正电影</summary>
      <form id="movie-match-search" data-media="${item.id}"><label>电影名<input id="movie-match-query" maxlength="128" placeholder="留空使用文件名"></label><label>年份<input id="movie-match-year" type="number" min="0" max="2100" value="${item.year || ''}" placeholder="0 表示不限年份"></label><button class="button secondary" type="submit">搜索候选</button></form>
      <form id="movie-match-id" data-media="${item.id}"><label>TMDB ID<input name="tmdb_id" type="number" min="1" max="2147483647" required placeholder="输入官网电影页面的数字 ID"></label><button class="button secondary" type="submit">使用此 ID 匹配</button></form>
      <p id="movie-match-error" class="form-error" hidden></p><div id="movie-match-results" aria-live="polite"></div>
    </details>`;
}
$('movie-detail').addEventListener('click', event => {
  const button = event.target.closest('[data-scrape-media]');
  if (button) startScraping({ media_id: Number(button.dataset.scrapeMedia), refresh: button.dataset.refresh === 'true' }, button);
  const candidate = event.target.closest('[data-match-tmdb]');
  if (candidate) startScraping({ media_id: movieDetailID, tmdb_id: Number(candidate.dataset.matchTmdb) }, candidate);
});
$('movie-detail').addEventListener('submit', async event => {
  const form = event.target;
  if (!['movie-match-search', 'movie-match-id'].includes(form.id)) return;
  event.preventDefault();
  const mediaID = Number(form.dataset.media);
  if (form.id === 'movie-match-id') {
    await startScraping({ media_id: mediaID, tmdb_id: Number(new FormData(form).get('tmdb_id')) }, event.submitter);
    return;
  }
  const revision = movieDetailRequest;
  const userID = state.user?.id;
  const active = () => movieDetailRequest === revision && state.user?.id === userID && $('movie-dialog').open;
  const query = new URLSearchParams();
  if ($('movie-match-query').value.trim()) query.set('q', $('movie-match-query').value.trim());
  if ($('movie-match-year').value !== '') query.set('year', $('movie-match-year').value);
  showError('movie-match-error');
  await busy(event.submitter, '正在搜索…', async () => {
    try {
      const result = await api(`/api/admin/media/${mediaID}/metadata/candidates?${query}`);
      if (!active()) return;
      $('movie-match-results').innerHTML = result.candidates.length ? result.candidates.map(item => `
        <article class="metadata-candidate"><strong>${escapeHTML(item.title)} (${item.year || '年份未知'})</strong><span>${escapeHTML(item.original_title)}</span><p>${escapeHTML(item.overview)}</p><button class="button secondary" type="button" data-match-tmdb="${item.id}">选择此电影 · ${item.id}</button></article>`).join('') : '<p class="field-help">未找到候选，可换一个关键词、将年份设为 0，或直接填写 TMDB ID。</p>';
    } catch (error) {
      if (active()) handleError(error, 'movie-match-error');
    }
  });
});
