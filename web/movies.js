'use strict';

const movies = { page: 1, total: 0, request: 0, timer: null };
function syncMovieLibraries() {
  const selected = $('movie-library-filter').value;
  const options = state.libraries.filter(item => item.type === 'movie');
  $('movie-library-filter').innerHTML = '<option value="">全部媒体库</option>' +
    options.map(item => `<option value="${item.id}">${escapeHTML(item.name)}</option>`).join('');
  if (options.some(item => String(item.id) === selected)) $('movie-library-filter').value = selected;
  $('scan-movies').disabled = !options.length;
}
function browseLibrary(id) {
  $('movie-library-filter').value = String(id);
  $('movie-search').value = '';
  movies.page = 1;
  switchSection('movies');
}
const movieDuration = seconds => seconds > 0 ? `${Math.max(1, Math.round(seconds / 60))} 分钟` : '时长待识别';
const movieSize = bytes => `${(bytes / (1024 ** 3)).toFixed(2)} GiB`;
const movieResolution = item => {
  if (item.width >= 3800 || item.height >= 2100) return '4K';
  if (item.width >= 1900 || item.height >= 1000) return '1080p';
  if (item.width >= 1260 || item.height >= 700) return '720p';
  return item.width ? `${item.width} × ${item.height}` : '待识别';
};

async function loadMovies() {
  if (!state.user) return;
  const userID = state.user.id;
  const request = ++movies.request;
  const query = new URLSearchParams({ page: String(movies.page), page_size: '24' });
  if ($('movie-library-filter').value) query.set('library_id', $('movie-library-filter').value);
  if ($('movie-search').value.trim()) query.set('q', $('movie-search').value.trim());
  showError('movie-error');
  $('movie-grid').setAttribute('aria-busy', 'true');
  try {
    const data = await api('/api/media?' + query);
    if (state.user?.id !== userID || request !== movies.request) return;
    movies.total = data.total;
    if (data.total > 0 && movies.page > Math.ceil(data.total / 24)) {
      movies.page = 1;
      return loadMovies();
    }
    $('movie-total').textContent = data.total;
    $('movie-grid').innerHTML = data.items.map(item => movieCard(item)).join('');
    bindPosterFallbacks($('movie-grid'));
    loadContinueWatching();
    $('movie-empty').hidden = data.items.length > 0;
    const filtering = Boolean(query.get('q') || query.get('library_id'));
    $('movie-empty-title').textContent = filtering ? '没有找到匹配的电影' : '珍藏，即将登场';
    $('movie-empty-description').textContent = filtering
      ? '试试其他关键词或媒体库，扫描中的电影也会陆续出现。'
      : '添加电影目录并完成扫描后，你的电影会展示在这里。';
    $('movie-page').textContent = `${data.total ? movies.page : 0} / ${Math.ceil(data.total / 24)}`;
    $('movie-prev').disabled = movies.page <= 1;
    $('movie-next').disabled = movies.page * 24 >= data.total;
    $('movie-pagination').hidden = data.total <= 24;
  } catch (error) {
    if (request === movies.request) handleError(error, 'movie-error');
  } finally {
    if (request === movies.request) $('movie-grid').removeAttribute('aria-busy');
  }
}
$('movie-search').addEventListener('input', () => {
  clearTimeout(movies.timer);
  movies.timer = setTimeout(() => { movies.page = 1; loadMovies(); }, 250);
});
$('movie-library-filter').addEventListener('change', () => { movies.page = 1; loadMovies(); });
$('movie-prev').addEventListener('click', () => { movies.page--; loadMovies(); });
$('movie-next').addEventListener('click', () => { movies.page++; loadMovies(); });
$('movie-reload').addEventListener('click', loadMovies);
$('movie-grid').addEventListener('click', event => {
  const card = event.target.closest('[data-movie]');
  if (card) openMovie(Number(card.dataset.movie));
});
