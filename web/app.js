'use strict';

const $ = id => document.getElementById(id);
const state = { section: null, user: null, csrf: '', setup: false, libraries: [], filter: 'all', editing: null, deleting: null, toastTimer: null };
const errors = {
  scrape_running: '已有刮削任务运行中，请完成后重试。',
  invalid_scrape: '刮削参数有误，请重新选择影片。',
  invalid_metadata_query: '请检查片名与年份，年份 0 表示不限。',
  scraper_unavailable: '刮削服务正在退出，请稍后重试。',
  invalid_metadata_settings: '凭据类型、语言或格式有误；切换类型时请填写新凭据。',
  tmdb_not_configured: '请先填写 TMDB 凭据。',
  tmdb_invalid_credential: 'TMDB 凭据无效，请检查是否选择了正确的类型并完整复制。',
  tmdb_unavailable: '无法连接 TMDB，请检查服务器网络后重试。',
  tmdb_rate_limited: 'TMDB 请求过于频繁，请稍后重试。',
  invalid_playback: '播放参数有误，或视频信息不完整，请重新扫描后重试。',
  media_unavailable: '视频文件不存在或当前无法读取。',
  invalid_progress: '观看进度无效，请重试。',
  scan_running: '媒体库正在扫描，请完成后再编辑或移除。',
  unsupported_library: '目前支持电影扫描，电视剧扫描将在后续版本开放。',
  scanner_unavailable: '扫描服务正在退出，请稍后重试。',
  invalid_filter: '筛选条件有误，请刷新后重试。',
  invalid_credentials: '用户名或密码不正确，请重新输入。', invalid_setup_token: '初始化令牌不正确，请检查服务器上的令牌文件。',
  already_initialized: '管理员已创建，请刷新页面后登录。', invalid_csrf: '页面验证已失效，请刷新后重试。',
  invalid_origin: '访问地址与服务器配置不一致，请使用服务器配置的访问地址。', unauthorized: '登录已过期，请重新登录。',
  forbidden: '你没有执行此操作的权限。', rate_limited: '尝试次数过多，请在 15 分钟后重试。',
  invalid_library_name: '媒体库名称应为 1–80 个字符。', invalid_library_type: '请选择电影或电视剧。',
  invalid_library_path: '请填写服务器上的绝对目录路径，例如 /mnt/media/movies。',
  library_path_unavailable: '目录不存在或无法读取，请检查服务器路径和权限。', library_path_exists: '该目录已经添加到媒体库，请勿重复配置。',
  not_found: '记录不存在，可能已被移除，请刷新后重试。', invalid_password: '密码不符合要求，请使用至少 12 字节的密码。',
  internal_error: '服务器处理请求时出现问题，请稍后重试。', request_timeout: '请求超时，请稍后重试。',
  body_too_large: '提交内容过长，请缩短后重试。', invalid_json: '提交内容有误，请检查后重试。',
};
const escapeHTML = text => String(text ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
const icon = name => `<svg class="icon" aria-hidden="true"><use href="#i-${name}"/></svg>`;

async function api(path, { method = 'GET', body, headers = {} } = {}) {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 16000);
  try {
    const requestHeaders = { ...headers };
    if (body !== undefined) requestHeaders['Content-Type'] = 'application/json';
    if (state.csrf && !['GET', 'HEAD'].includes(method)) requestHeaders['X-CSRF-Token'] = state.csrf;
    const response = await fetch(path, { method, headers: requestHeaders, credentials: 'same-origin', signal: controller.signal, body: body === undefined ? undefined : JSON.stringify(body) });
    const data = await response.json();
    if (!response.ok) {
      const error = new Error(errors[data.error?.code] || '请求失败，请稍后重试。');
      error.status = response.status;
      error.code = data.error?.code;
      throw error;
    }
    return data;
  } catch (error) {
    if (error.name === 'AbortError') throw new Error('连接超时，请检查服务器状态后重试。');
    if (error instanceof TypeError) throw new Error('无法连接服务器，请检查网络后重试。');
    throw error;
  } finally { clearTimeout(timer); }
}
function showError(id, message = '') { $(id).textContent = message; $(id).hidden = !message; }
function notify(message, error = false) {
  clearTimeout(state.toastTimer);
  $('toast').textContent = message;
  $('toast').classList.toggle('error', error);
  $('toast').hidden = false;
  state.toastTimer = setTimeout(() => { $('toast').hidden = true; }, 5000);
}
function handleError(error, target) {
  if (error.status === 401 && state.user) {
    document.querySelectorAll('dialog[open]').forEach(dialog => dialog.close());
    showAuth(false);
    notify('登录已过期，请重新登录。', true);
  } else if (target) showError(target, error.message);
  else notify(error.message, true);
}
async function busy(button, label, action) {
  const previousLabel = button.getAttribute('aria-label');
  button.disabled = true;
  button.setAttribute('aria-busy', 'true');
  button.setAttribute('aria-label', label);
  try { return await action(); }
  finally {
    button.disabled = false;
    button.removeAttribute('aria-busy');
    if (previousLabel === null) button.removeAttribute('aria-label');
    else button.setAttribute('aria-label', previousLabel);
  }
}
function showAuth(setup) {
  resetMetadataSettings();
  stopScrapePolling();
  resetScrapeAdmin();
  movieDetailRequest++;
  if (state.section === 'player') closePlayer();
  state.section = null;
  state.setup = setup; state.user = null; state.csrf = ''; state.libraries = []; stopScanPolling();
  $('boot').hidden = true; $('app').hidden = true; $('auth-screen').hidden = false;
  $('auth-form').reset(); showError('auth-error');
  $('setup-token-label').hidden = !setup; $('setup-token').required = setup;
  $('confirm-label').hidden = !setup; $('confirm-password').required = setup;
  $('password').type = 'password'; $('toggle-password').textContent = '显示'; $('toggle-password').setAttribute('aria-label', '显示密码');
  $('password').autocomplete = setup ? 'new-password' : 'current-password';
  $('password').removeAttribute('minlength');
  $('auth-label').textContent = setup ? 'MAKE YOURSELF AT HOME' : 'WELCOME BACK';
  $('auth-title').textContent = setup ? '开启你的私人空间' : '欢迎回来';
  $('auth-description').textContent = setup ? '创建首位管理员，开始整理你的媒体珍藏。' : '登录 Dinosaur，管理你的私人媒体空间。';
  $('auth-submit-text').textContent = setup ? '创建管理员' : '登录';
  $('auth-footnote').textContent = setup ? '初始化只需一次，密码至少 12 字节' : '私人服务器，仅限已授权用户访问';
  document.title = `${setup ? '初始化' : '登录'} · Dinosaur`;
  history.replaceState(null, '', setup ? '/setup' : '/login');
}
async function showApp(data) {
  state.user = data.user; state.csrf = data.csrf_token; state.filter = 'all';
  $('auth-form').reset(); $('boot').hidden = true; $('auth-screen').hidden = true; $('app').hidden = false;
  $('user-name').textContent = data.user.username;
  $('user-avatar').textContent = data.user.username.slice(0, 1).toUpperCase();
  $('user-role').textContent = data.user.role === 'admin' ? '管理员' : '普通用户';
  const admin = data.user.role === 'admin';
  $('metadata-nav').hidden = !admin;
  $('scrape-admin-nav').hidden = !admin;
  $('scrape-movies').hidden = !admin;
  $('scan-movies').hidden = !admin;
  $('add-library').hidden = !admin; $('empty-add').hidden = !admin;
  $('library-hint').textContent = admin ? '添加服务器上的媒体目录，配置会自动保存。媒体扫描会自动开始，也可以手动重新扫描。' : '查看管理员配置的媒体库。添加、编辑和移除目录需要管理员权限。';
  $('library-search').value = '';
  document.querySelectorAll('[data-filter]').forEach(button => { const active = button.dataset.filter === 'all'; button.classList.toggle('active', active); button.setAttribute('aria-pressed', String(active)); });
  restorePage(true);
  health(); await loadLibraries();
}
async function health() {
  try { await api('/health/ready'); $('health-dot').className = 'status-dot online'; $('health-text').textContent = '服务器运行正常'; }
  catch { $('health-dot').className = 'status-dot offline'; $('health-text').textContent = '服务器连接异常'; }
}
$('auth-form').addEventListener('submit', async event => {
  event.preventDefault(); showError('auth-error');
  if (state.setup && $('password').value !== $('confirm-password').value) { showError('auth-error', '两次输入的密码不一致。'); return; }
  if (state.setup && new TextEncoder().encode($('password').value).length < 12) { showError('auth-error', '密码至少需要 12 字节。'); return; }
  const body = { username: $('username').value.trim(), password: $('password').value };
  await busy(event.submitter, state.setup ? '正在创建…' : '正在登录…', async () => {
    try {
      if (state.setup) {
        await api('/api/setup', { method: 'POST', body, headers: { 'X-Setup-Token': $('setup-token').value.trim() } });
        showAuth(false); $('username').value = body.username; notify('管理员已创建，请使用刚才的密码登录。');
      } else await showApp(await api('/api/auth/login', { method: 'POST', body }));
    } catch (error) { showError('auth-error', error.message); }
  });
});
$('toggle-password').addEventListener('click', () => {
  const visible = $('password').type === 'password'; $('password').type = visible ? 'text' : 'password';
  $('toggle-password').textContent = visible ? '隐藏' : '显示'; $('toggle-password').setAttribute('aria-label', visible ? '隐藏密码' : '显示密码');
});
$('logout').addEventListener('click', async () => {
  await busy($('logout'), '…', async () => { try { await api('/api/auth/logout', { method: 'POST' }); showAuth(false); notify('已安全退出。'); } catch (error) { handleError(error); } });
});
document.querySelectorAll('[data-section]').forEach(button => button.addEventListener('click', () => switchSection(button.dataset.section)));
$('password-form').addEventListener('submit', async event => {
  event.preventDefault(); showError('password-error');
  const data = new FormData(event.target);
  if (data.get('new_password') !== data.get('confirm_password')) { showError('password-error', '两次输入的新密码不一致。'); return; }
  await busy(event.submitter, '正在更新…', async () => {
    try { await api('/api/auth/password', { method: 'POST', body: { current_password: data.get('current_password'), new_password: data.get('new_password') } }); event.target.reset(); showAuth(false); notify('密码已更新，请重新登录。'); }
    catch (error) { handleError(error, 'password-error'); }
  });
});
async function boot() {
  $('retry-boot').hidden = true; $('boot').querySelector('p').textContent = '正在连接你的媒体空间…';
  try {
    const setup = await api('/api/setup');
    if (!setup.initialized) { showAuth(true); return; }
    try { await showApp(await api('/api/auth/me')); }
    catch (error) { if (error.status === 401) showAuth(false); else throw error; }
  } catch (error) { $('boot').querySelector('p').textContent = error.message; $('retry-boot').hidden = false; }
}
$('retry-boot').addEventListener('click', boot);
document.addEventListener('DOMContentLoaded', boot);
