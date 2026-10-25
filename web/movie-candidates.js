'use strict';

function candidateCards(candidates, mediaID, compact = false) {
  return candidates.slice(0, 5).map((item, index) => `<article class="metadata-candidate ${compact ? 'candidate-compact' : ''}">
    <strong>${index + 1}. ${escapeHTML(item.title)} (${item.year || '年份未知'})</strong>
    ${item.original_title && item.original_title !== item.title ? `<span>${escapeHTML(item.original_title)}</span>` : ''}
    ${!compact && item.overview ? `<p>${escapeHTML(item.overview)}</p>` : ''}
    <button class="${compact ? 'text-button' : 'button secondary'}" type="button" data-select-candidate="${item.id}" data-candidate-media="${mediaID}" ${scraping.active ? 'disabled' : ''}>选择此电影 · ${item.id}</button>
  </article>`).join('');
}
function candidatePreview(mediaID, candidates = [], message = '', compact = false) {
  if (!candidates.length && !message.split(',').includes('ambiguous')) return '';
  return `<section class="candidate-preview" data-candidates-for="${mediaID}" data-compact="${compact}">
    <h3>相似候选（最多 5 部）</h3><div class="candidate-list">
      ${candidates.length ? candidateCards(candidates, mediaID, compact) : '<button class="text-button" type="button" data-load-candidates>加载前五个候选</button>'}
    </div></section>`;
}
async function loadCandidatePreview(element) {
  const userID = state.user?.id;
  const mediaID = Number(element.dataset.candidatesFor);
  const active = () => element.isConnected && state.user?.id === userID;
  element.setAttribute('aria-busy', 'true');
  element.querySelector('.candidate-list').textContent = '正在获取前五个候选…';
  try {
    const result = await api(`/api/admin/media/${mediaID}/metadata/candidates`);
    if (!active()) return;
    element.querySelector('.candidate-list').innerHTML = result.candidates.length
      ? candidateCards(result.candidates, mediaID, element.dataset.compact === 'true')
      : '<p class="field-help">没有找到候选，请使用“匹配 / 更正电影”调整片名或年份。</p>';
  } catch (error) {
    if (!active()) return;
    if (error.status === 401) handleError(error);
    else element.querySelector('.candidate-list').innerHTML = `<p class="form-error">${escapeHTML(error.message)}</p><button class="text-button" type="button" data-load-candidates>重新加载候选</button>`;
  } finally {
    if (active()) element.removeAttribute('aria-busy');
  }
}
function loadLegacyMovieCandidates(item) {
  if (state.user?.role !== 'admin' || item.metadata?.candidates?.length || !item.metadata?.message?.split(',').includes('ambiguous')) return;
  const preview = $('movie-detail').querySelector('[data-candidates-for]');
  if (preview) loadCandidatePreview(preview);
}
document.addEventListener('click', event => {
  const load = event.target.closest('[data-load-candidates]');
  if (load && state.user?.role === 'admin') loadCandidatePreview(load.closest('[data-candidates-for]'));
  const button = event.target.closest('[data-select-candidate]');
  if (button && state.user?.role === 'admin') startScraping({ media_id: Number(button.dataset.candidateMedia), tmdb_id: Number(button.dataset.selectCandidate) }, button);
});
