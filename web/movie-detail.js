'use strict';

let movieDetailRequest = 0;
let movieDetailID = null;
let movieDetailOrigin = 'movies';
function movieDetailHeader() {
  return `<header class="detail-page-header"><button class="icon-button detail-back" type="button" aria-label="返回上一页" title="返回"><svg class="icon" aria-hidden="true"><use href="#i-arrow"/></svg></button><span>电影详情</span></header>`;
}
function moviePlayActions(item) {
  const resume = item.position > 0 && !item.watched;
  return `<div class="movie-play-actions"><button class="button primary" data-play-movie="${item.id}">${resume ? '继续观看 · ' + playerClock(item.position) : '播放电影'}</button>${resume ? `<button class="button secondary" data-play-movie="${item.id}" data-beginning="true">从头播放</button>` : ''}<button class="text-button" data-watched-movie="${item.id}" data-watched="${item.watched}">${item.watched ? '标记为未看' : '标记为已看'}</button></div>`;
}
async function openMovie(id, { replace = false } = {}) {
  if (state.section !== 'detail') {
    movieDetailOrigin = replace ? history.state?.detailOrigin || 'movies' : state.section === 'scraping' ? 'scraping' : 'movies';
  }
  movieDetailID = id;
  const playbackSaved = switchSection('detail', id, { replace });
  const request = ++movieDetailRequest;
  const userID = state.user?.id;
  const active = () => state.user?.id === userID && request === movieDetailRequest && state.section === 'detail';
  $('movie-detail').innerHTML = movieDetailHeader() + '<p class="detail-loading" role="status">正在读取电影信息…</p>';
  try {
    await playbackSaved;
    if (!active()) return;
    const { item } = await api(`/api/media/${id}`);
    if (!active()) return;
    document.title = `${item.title} · Dinosaur`;
    const streams = item.tracks.map(track => `<li><strong>${track.type === 'audio' ? '音轨' : '字幕'} ${track.index}</strong><span>${escapeHTML([track.language, track.codec, track.channels ? `${track.channels} 声道` : '', track.title].filter(Boolean).join(' · '))}</span></li>`).join('');
    $('movie-detail').innerHTML = `${movieDetailHeader()}
      <div class="detail-layout">
        <aside class="detail-artwork"><div class="detail-poster-frame"><span class="detail-poster-placeholder">${icon('film')}</span>${item.poster_url ? `<img class="detail-poster" src="${escapeHTML(item.poster_url)}" alt="${escapeHTML(item.title)} 海报">` : ''}</div></aside>
        <article class="detail-copy"><span class="small-label">${escapeHTML(item.library_name)}</span>
          <h1 id="movie-detail-title">${escapeHTML(item.title)}</h1>
          <p class="detail-meta">${item.year || '年份未知'}<span>·</span>${movieDuration(item.duration)}<span>·</span>${movieResolution(item)}</p>
          ${moviePlayActions(item)}
          ${movieMetadataHTML(item)}
          ${item.probe_error ? '<p class="form-error">视频信息暂未读取完整，可以重新扫描。电影文件已被索引。</p>' : ''}
          <details class="movie-technical"><summary>媒体信息</summary>
            <dl class="movie-facts"><div><dt>视频编码</dt><dd>${escapeHTML(item.video_codec.toUpperCase() || '待识别')}</dd></div><div><dt>分辨率</dt><dd>${item.width && item.height ? `${item.width} × ${item.height}` : '待识别'}</dd></div><div><dt>文件大小</dt><dd>${movieSize(item.size)}</dd></div></dl>
            ${streams ? `<h3>音轨与字幕</h3><ul class="movie-tracks">${streams}</ul>` : ''}
            <div class="movie-file"><h3>文件信息</h3><p>${escapeHTML(item.filename)}</p><p>容器：${escapeHTML(item.format || '待识别')}</p></div>
          </details>
        </article>
      </div>`;
    loadLegacyMovieCandidates(item);
    $('movie-detail').querySelector('img')?.addEventListener('error', event => { event.currentTarget.hidden = true; });
  } catch (error) {
    if (!active()) return;
    if (error.status === 401) handleError(error);
    else $('movie-detail').innerHTML = `${movieDetailHeader()}<p class="detail-loading form-error" role="alert">${escapeHTML(error.message)}</p><button class="button secondary detail-retry" type="button">重新加载</button>`;
  }
}
$('movie-detail').addEventListener('click', event => {
  if (event.target.closest('.detail-back')) {
    if (history.state?.detailBack) history.back();
    else switchSection(movieDetailOrigin, null, { replace: true });
  }
  if (event.target.closest('.detail-retry')) openMovie(movieDetailID, { replace: true });
});
