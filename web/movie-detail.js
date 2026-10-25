'use strict';

let movieDetailRequest = 0;
let movieDetailID = null;
async function openMovie(id) {
  const request = ++movieDetailRequest;
  const userID = state.user?.id;
  movieDetailID = id;
  $('movie-detail').innerHTML = '<p class="loading-message">正在读取电影信息…</p>';
  if (!$('movie-dialog').open) $('movie-dialog').showModal();
  try {
    const { item } = await api(`/api/media/${id}`);
    if (state.user?.id !== userID || request !== movieDetailRequest || !$('movie-dialog').open) return;
    const streams = item.tracks.map(track => `<li><strong>${track.type === 'audio' ? '音轨' : '字幕'} ${track.index}</strong><span>${escapeHTML([track.language, track.codec, track.channels ? `${track.channels} 声道` : '', track.title].filter(Boolean).join(' · '))}</span></li>`).join('');
    $('movie-detail').innerHTML = `
      ${item.poster_url ? `<img class="detail-poster" src="${escapeHTML(item.poster_url)}" alt="${escapeHTML(item.title)} 封面">` : ''}
      <span class="small-label">${escapeHTML(item.library_name)}</span>
      <h2 id="movie-detail-title">${escapeHTML(item.title)}</h2>
      <p class="detail-meta">${item.year || '年份未知'}<span>·</span>${movieDuration(item.duration)}<span>·</span>${movieResolution(item)}</p>
      ${movieMetadataHTML(item)}
      ${item.probe_error ? '<p class="form-error">视频信息暂未读取完整，可以重新扫描。电影文件已被索引。</p>' : ''}
      <dl class="movie-facts"><div><dt>视频编码</dt><dd>${escapeHTML(item.video_codec.toUpperCase() || '待识别')}</dd></div><div><dt>分辨率</dt><dd>${item.width && item.height ? `${item.width} × ${item.height}` : '待识别'}</dd></div><div><dt>文件大小</dt><dd>${movieSize(item.size)}</dd></div></dl>
      ${streams ? `<h3>音轨与字幕</h3><ul class="movie-tracks">${streams}</ul>` : ''}
      <details class="movie-file"><summary>文件信息</summary><p>${escapeHTML(item.filename)}</p><p>容器：${escapeHTML(item.format || '待识别')}</p></details>
      <div class="movie-play-actions"><button class="button primary" data-play-movie="${item.id}">${item.position > 0 && !item.watched ? '继续观看 · ' + playerClock(item.position) : '播放电影'}</button>${item.position > 0 && !item.watched ? `<button class="button secondary" data-play-movie="${item.id}" data-beginning="true">从头播放</button>` : ''}<button class="text-button" data-watched-movie="${item.id}" data-watched="${item.watched}">${item.watched ? '标记为未看' : '标记为已看'}</button></div>`;
    $('movie-detail').querySelector('img')?.addEventListener('error', event => event.currentTarget.remove());
  } catch (error) {
    if (state.user?.id !== userID) return;
    if (error.status === 401) handleError(error);
    else if (request === movieDetailRequest) $('movie-detail').innerHTML = `<p class="form-error">${escapeHTML(error.message)}</p>`;
  }
}
