'use strict';

async function loadLibraries(quiet = false) {
  if (!state.user) return;
  if (!quiet) {
    showError('library-load-error');
    $('empty-state').hidden = true;
    $('library-list').innerHTML = '<p class="loading-message">正在读取媒体库…</p>';
  }
  try {
    const endpoint = state.user.role === 'admin' ? '/api/admin/libraries' : '/api/libraries';
    const result = await api(endpoint);
    if (!state.user) return;
    state.libraries = result.libraries;
    renderLibraries();
    syncMovieLibraries();
    syncSeriesLibraries();
    syncScrapeAdminLibraries();
    scheduleScanPolling();
    if (state.user.role === 'admin') {
      loadScrapeJobs();
      loadSeriesScrapeJobs();
    }
  } catch (error) {
    if (!quiet) $('library-list').replaceChildren();
    handleError(error, 'library-load-error');
  }
}

function renderLibraries() {
  const admin = state.user?.role === 'admin';
  const query = $('library-search').value.trim().toLocaleLowerCase();
  const filtered = state.libraries.filter(item =>
    (state.filter === 'all' || state.filter === item.type) && item.name.toLocaleLowerCase().includes(query));
  $('nav-count').textContent = state.libraries.length;
  $('all-count').textContent = state.libraries.length;
  $('library-list').innerHTML = filtered.map(item => {
    const scanning = ['queued', 'scanning'].includes(item.scan?.status);
    const actions = admin ? `<div class="card-actions">
      <button class="icon-button" data-action="edit" data-id="${item.id}" aria-label="编辑 ${escapeHTML(item.name)}" title="编辑" ${scanning ? 'disabled' : ''}>${icon('edit')}</button>
      <button class="icon-button" data-action="delete" data-id="${item.id}" aria-label="移除 ${escapeHTML(item.name)}" title="移除" ${scanning ? 'disabled' : ''}>${icon('trash')}</button></div>` : '';
    return `<article class="library-card">
      <div class="card-top"><span class="library-type-icon ${item.type === 'tv' ? 'tv' : ''}">${icon(item.type === 'tv' ? 'tv' : 'film')}</span><span class="type-tag">${item.type === 'tv' ? '电视剧' : '电影'}</span></div>
      <h2>${escapeHTML(item.name)}</h2>
      <p class="card-path" title="${escapeHTML(item.root_path || '')}">${escapeHTML(item.root_path || '目录由管理员管理')}</p>
      <div class="card-bottom"><span class="card-status ${admin && !item.available ? 'unavailable' : ''}"><span class="status-dot"></span>${admin ? (item.available ? '目录可访问' : '目录暂不可用') : '已配置'}</span>${actions}</div>
      <div class="library-browse-row"><button class="text-button" data-action="browse" data-id="${item.id}">${item.type === 'tv' ? '查看剧集' : '查看电影'} <span>${item.media_count || 0}</span></button>${admin ? `<button class="text-button" data-action="scan" data-id="${item.id}" ${scanning ? 'disabled' : ''}>${scanning ? '正在扫描…' : '重新扫描'}</button>` : ''}</div>
      ${admin && item.scan ? `<p class="scan-summary">${escapeHTML(scanDescription(item.scan))}</p>` : ''}
      </article>`;
  }).join('');
  $('empty-state').hidden = filtered.length > 0;
  const none = !state.libraries.length;
  $('empty-title').textContent = none ? '你的媒体故事，从这里开始' : '没有找到匹配的媒体库';
  $('empty-description').textContent = none
    ? (admin ? '添加第一个媒体库，把收藏的电影和剧集整理在一起。' : '管理员尚未添加媒体库，请稍后再来看看。')
    : '换一个关键词，或选择其他媒体类型试试。';
  $('empty-add').hidden = !admin || !none;
}

function openLibrary(item = null) {
  state.editing = item?.id ?? null;
  $('library-form').reset();
  showError('library-form-error');
  $('library-dialog-title').textContent = item ? '编辑媒体库' : '添加媒体库';
  $('library-name').value = item?.name ?? '';
  $('library-path').value = item?.root_path ?? '';
  document.querySelector(`input[name="library_type"][value="${item?.type ?? 'movie'}"]`).checked = true;
  $('library-dialog').showModal();
}

document.querySelectorAll('[data-filter]').forEach(button => button.addEventListener('click', () => {
  state.filter = button.dataset.filter;
  document.querySelectorAll('[data-filter]').forEach(tab => {
    const active = tab === button;
    tab.classList.toggle('active', active);
    tab.setAttribute('aria-pressed', String(active));
  });
  renderLibraries();
}));
$('library-search').addEventListener('input', renderLibraries);
$('add-library').addEventListener('click', () => openLibrary());
$('empty-add').addEventListener('click', () => openLibrary());
document.querySelectorAll('[data-close]').forEach(button =>
  button.addEventListener('click', () => $(button.dataset.close).close()));

$('library-list').addEventListener('click', async event => {
  const button = event.target.closest('[data-action]');
  if (!button) return;
  const item = state.libraries.find(library => library.id === Number(button.dataset.id));
  if (!item) return;
  switch (button.dataset.action) {
    case 'edit': openLibrary(item); break;
    case 'browse': if (item.type === 'tv') browseSeriesLibrary(item.id); else browseLibrary(item.id); break;
    case 'scan': await scanLibraries([item]); break;
    case 'delete':
      state.deleting = item.id;
      $('delete-name').textContent = item.name;
      showError('delete-error');
      $('delete-dialog').showModal();
  }
});

$('library-form').addEventListener('submit', async event => {
  event.preventDefault();
  showError('library-form-error');
  const id = state.editing;
  const body = {
    name: $('library-name').value.trim(),
    type: document.querySelector('input[name="library_type"]:checked').value,
    root_path: $('library-path').value,
  };
  await busy($('save-library'), '正在保存…', async () => {
    try {
      await api(id ? `/api/admin/libraries/${id}` : '/api/admin/libraries', { method: id ? 'PUT' : 'POST', body });
      $('library-dialog').close();
      notify(id ? '媒体库配置已更新。' : '媒体库已添加，扫描将自动开始。');
      await loadLibraries();
    } catch (error) { handleError(error, 'library-form-error'); }
  });
});

$('confirm-delete').addEventListener('click', async () => {
  const id = state.deleting;
  await busy($('confirm-delete'), '正在移除…', async () => {
    try {
      await api(`/api/admin/libraries/${id}`, { method: 'DELETE' });
      $('delete-dialog').close();
      notify('媒体库已移除，原始文件已保留。');
      await loadLibraries();
    } catch (error) { handleError(error, 'delete-error'); }
  });
});
