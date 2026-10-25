'use strict';

let metadataRevision = 0;
function resetMetadataSettings() {
  metadataRevision++;
  $('tmdb-form').reset();
  $('tmdb-fields').disabled = true;
  $('tmdb-state').textContent = '正在读取配置…';
  $('tmdb-message').textContent = '';
  showError('tmdb-error');
}
function displayMetadataSettings(settings) {
  $('tmdb-type').value = settings.credential_type;
  $('tmdb-language').value = settings.language;
  $('tmdb-credential').value = '';
  $('tmdb-credential').placeholder = settings.configured ? '已配置；留空保留，输入新凭据可替换' : '粘贴 TMDB 凭据';
  $('tmdb-state').textContent = settings.configured ? '已配置凭据' : '尚未配置凭据';
  $('tmdb-clear').disabled = !settings.configured;
}
async function loadMetadataSettings() {
  resetMetadataSettings();
  const revision = metadataRevision;
  const userID = state.user?.id;
  const active = () => revision === metadataRevision && state.user?.id === userID;
  $('tmdb-reload').hidden = true;
  try {
    const settings = await api('/api/admin/metadata/tmdb');
    if (!active()) return;
    displayMetadataSettings(settings);
    $('tmdb-fields').disabled = false;
    loadScrapeJobs();
  } catch (error) {
    if (!active()) return;
    $('tmdb-state').textContent = '配置读取失败';
    $('tmdb-reload').hidden = false;
    handleError(error, 'tmdb-error');
  }
}
function metadataSettingsInput() {
  return {
    credential_type: $('tmdb-type').value,
    language: $('tmdb-language').value,
    credential: $('tmdb-credential').value.trim(),
  };
}
async function submitMetadataSettings(action) {
  const revision = metadataRevision;
  const userID = state.user?.id;
  const active = () => revision === metadataRevision && state.user?.id === userID;
  const body = metadataSettingsInput();
  $('tmdb-fields').disabled = true;
  $('tmdb-message').textContent = action === 'test' ? '正在测试 TMDB 连接…' : '正在保存…';
  showError('tmdb-error');
  try {
    const path = '/api/admin/metadata/tmdb' + (action === 'test' ? '/test' : '');
    const result = await api(path, {
      method: action === 'test' ? 'POST' : action === 'clear' ? 'DELETE' : 'PUT',
      body: action === 'clear' ? undefined : body,
    });
    if (!active()) return;
    if (action !== 'test') displayMetadataSettings(result);
    $('tmdb-message').textContent = action === 'test' ? '连接成功，凭据有效。测试不会保存配置。' : action === 'clear' ? '凭据已移除。' : '配置已保存。';
  } catch (error) {
    if (!active()) return;
    $('tmdb-message').textContent = '';
    handleError(error, 'tmdb-error');
  } finally {
    if (active()) $('tmdb-fields').disabled = false;
  }
}
$('tmdb-form').addEventListener('submit', event => {
  event.preventDefault();
  submitMetadataSettings('save');
});
$('tmdb-test').addEventListener('click', () => submitMetadataSettings('test'));
$('tmdb-clear').addEventListener('click', () => submitMetadataSettings('clear'));
$('tmdb-reload').addEventListener('click', loadMetadataSettings);
