'use strict';

function movieCard(movie, resume = false) {
  const action = resume ? `data-play-movie="${movie.id}"` : `data-movie="${movie.id}"`;
  const label = resume ? '继续观看' : '查看';
  return `<button class="movie-card" ${action} aria-label="${label} ${escapeHTML(movie.title)}">
    <span class="movie-artwork">
      <span class="movie-placeholder">${icon('film')}<span>DINOSAUR COLLECTION</span></span>
      ${movie.poster_url ? `<img class="movie-poster" src="${escapeHTML(movie.poster_url)}" alt="${escapeHTML(movie.title)} 封面" loading="lazy">` : ''}
      ${itemMetadataBadge(movie)}
      <span class="movie-quality">${movieResolution(movie)}</span>
      ${movie.watched ? '<span class="movie-watched">已看</span>' : ''}
      <span class="movie-artwork-caption">${escapeHTML(movie.library_name)}</span>
      ${movie.position > 0 && movie.duration > 0 && !movie.watched ? `<progress class="movie-watch-progress" value="${movie.position}" max="${movie.duration}" aria-label="观看进度"></progress>` : ''}
    </span>
    <span class="movie-title">${escapeHTML(movie.title)}</span>
    <span class="movie-meta">${movie.year || '年份未知'}<span>·</span>${resume ? `已看 ${playerClock(movie.position)}` : movieDuration(movie.duration)}</span>
  </button>`;
}
function bindPosterFallbacks(root) {
  root.querySelectorAll('img').forEach(image => image.addEventListener('error', () => image.remove()));
}
async function loadContinueWatching() {
  const userID = state.user?.id;
  try {
    const data = await api('/api/playback/continue');
    if (state.user?.id !== userID) return;
    $('continue-section').hidden = !data.items.length;
    $('continue-grid').innerHTML = data.items.map(item => movieCard(item, true)).join('');
    bindPosterFallbacks($('continue-grid'));
  } catch (error) {
    $('continue-section').hidden = true;
    if (error.status === 401) handleError(error);
  }
}

function itemMetadataBadge(movie) {
  const status = movie.metadata?.status;
  const label = status === 'needs_review' ? '待确认匹配' : status === 'failed' ? '刮削失败' : '';
  return label ? `<span class="movie-metadata-badge">${label}</span>` : '';
}
