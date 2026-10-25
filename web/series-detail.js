'use strict';

function seriesHeader() {
  return `<header class="detail-page-header"><button class="icon-button series-back" type="button" aria-label="返回剧集列表" title="返回剧集列表">${icon('arrow')}</button><span>剧集详情</span></header>`;
}
async function openSeries(id, { replace = false } = {}) {
  if (series.id !== id) series.season = null;
  series.id = id;
  const saved = switchSection('series-detail', id, { replace });
  const request = ++series.detailRequest;
  const userID = state.user?.id;
  const active = () => state.user?.id === userID && request === series.detailRequest && state.section === 'series-detail';
  $('series-detail').innerHTML = seriesHeader() + '<p class="detail-loading" role="status">正在读取剧集信息…</p>';
  try {
    await saved;
    if (!active()) return;
    const detail = await api(`/api/series/${id}`);
    if (!active()) return;
    series.detail = detail;
    const item = detail.series;
    document.title = `${item.title} · Dinosaur`;
    $('series-detail').innerHTML = `${seriesHeader()}<div class="detail-layout">
      <aside class="detail-artwork"><div class="detail-poster-frame"><span class="detail-poster-placeholder">${icon('tv')}</span>${item.poster_url ? `<img class="detail-poster" src="${escapeHTML(item.poster_url)}" alt="${escapeHTML(item.title)} 海报">` : ''}</div></aside>
      <article class="detail-copy"><span class="small-label">${escapeHTML(item.library_name)}</span><h1>${escapeHTML(item.title)}</h1>
        <p class="detail-meta">${item.year || '年份未知'} · ${item.season_count} 季 · ${item.episode_count} 个文件 · 已看 ${item.watched_count}</p>
        ${detail.resume ? `<div class="movie-play-actions"><button class="button primary" data-play-movie="${detail.resume.id}">${detail.resume.position > 0 && !detail.resume.watched ? '继续观看' : '播放'} · ${escapeHTML(detail.resume.title)}</button></div>` : ''}
        ${seriesMetadataHTML(item)}
        <div class="series-seasons"><label>选择季度<select id="series-season">${detail.seasons.map(season => `<option value="${season.number}">${season.number === 0 ? '特别篇' : `第 ${season.number} 季`} · ${season.episodes.length} 个文件</option>`).join('')}</select></label><div id="series-episodes" class="series-episodes"></div></div>
      </article></div>`;
    if (!detail.seasons.some(item => item.number === series.season)) series.season = detail.resume?.season_number ?? detail.seasons[0]?.number;
    $('series-season').value = String(series.season);
    renderSeriesEpisodes();
    $('series-season').addEventListener('change', () => { series.season = Number($('series-season').value); renderSeriesEpisodes(); });
    bindPosterFallbacks($('series-detail'));
    if (state.user.role === 'admin') loadSeriesScrapeJobs();
  } catch (error) {
    if (!active()) return;
    if (error.status === 401) handleError(error);
    else $('series-detail').innerHTML = `${seriesHeader()}<p class="detail-loading form-error">${escapeHTML(error.message)}</p><button class="button secondary series-retry">重新加载</button>`;
  }
}
function renderSeriesEpisodes() {
  const season = series.detail.seasons.find(item => item.number === series.season);
  $('series-episodes').innerHTML = (season?.episodes || []).map(item => `<article class="series-episode">
    <div><h2>${escapeHTML(item.title)}</h2><p class="muted">${movieDuration(item.duration)} · ${movieResolution(item)}${item.watched ? ' · 已看' : item.position > 0 ? ` · 已看 ${playerClock(item.position)}` : ''}</p>
      ${item.position > 0 && item.duration > 0 ? `<progress value="${item.position}" max="${item.duration}" aria-label="观看进度"></progress>` : ''}
      <details><summary>媒体信息${item.probe_error ? ' · 信息读取失败，可重新扫描' : ''}</summary><p class="field-help">${escapeHTML(item.filename)}<br>${escapeHTML(item.video_codec)} · ${movieSize(item.size)}</p><ul class="movie-tracks">${item.tracks.map(track => `<li>${escapeHTML([track.type, track.codec, track.language, track.title].filter(Boolean).join(' · '))}</li>`).join('')}</ul></details>
    </div><div class="series-episode-actions"><button class="button secondary" data-play-movie="${item.id}">${item.position > 0 && !item.watched ? '继续观看' : '播放'}</button>
      ${item.position > 0 && !item.watched ? `<button class="text-button" data-play-movie="${item.id}" data-beginning="true">从头播放</button>` : ''}
      <button class="text-button" data-series-watched="${item.id}" data-watched="${item.watched}">${item.watched ? '标记未看' : '标记已看'}</button></div></article>`).join('');
}
$('series-detail').addEventListener('click', async event => {
  if (event.target.closest('.series-back')) switchSection('series');
  if (event.target.closest('.series-retry')) openSeries(series.id, { replace: true });
  const button = event.target.closest('[data-series-watched]');
  if (button) {
    const id = series.id;
    const userID = state.user?.id;
    await busy(button, '正在保存…', async () => {
      try {
        await api(`/api/playback/${button.dataset.seriesWatched}/progress`, { method: 'PUT', body: { watched: button.dataset.watched !== 'true' } });
        if (state.user?.id === userID && state.section === 'series-detail' && series.id === id) await openSeries(id, { replace: true });
      } catch (error) { if (state.user?.id === userID) handleError(error); }
    });
  }
});
