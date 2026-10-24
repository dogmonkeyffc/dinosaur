'use strict';

let scanPollTimer;
function stopScanPolling() {
  clearTimeout(scanPollTimer);
}
function scanDescription(scan) {
  const progress = `${scan.processed} / ${scan.discovered}`;
  switch (scan.status) {
    case 'queued': return '扫描已排队，稍后开始';
    case 'scanning': return `正在扫描 · ${progress}`;
    case 'completed': return `扫描完成 · ${scan.processed} 个视频${scan.errors ? ` · ${scan.errors} 项信息或封面读取失败` : ''}`;
    case 'failed': return '扫描失败，请检查目录访问权限后重试';
    default: return '扫描已中断，可以重新扫描';
  }
}
function scheduleScanPolling() {
  stopScanPolling();
  const active = state.libraries.filter(item => ['queued', 'scanning'].includes(item.scan?.status));
  $('movie-scan-status').hidden = active.length === 0;
  $('movie-scan-status').textContent = active.map(item => `${item.name}：${scanDescription(item.scan)}`).join('；');
  if (!active.length || !state.user) return;
  scanPollTimer = setTimeout(async () => {
    await loadLibraries(true);
    if (state.user && !$('movies-section').hidden) await loadMovies();
  }, 2000);
}
async function scanLibraries(libraries) {
  if (!libraries.length) { notify('请先在媒体库设置中添加电影目录。'); return; }
  try {
    for (const item of libraries) {
      try { await api(`/api/admin/libraries/${item.id}/scan`, { method: 'POST' }); }
      catch (error) { if (error.code !== 'scan_running') throw error; }
    }
    notify('扫描已开始，电影列表会自动更新。');
    await loadLibraries(true);
  } catch (error) { handleError(error); }
}
$('scan-movies').addEventListener('click', () => busy($('scan-movies'), '正在开始扫描…', async () => {
  const selected = Number($('movie-library-filter').value);
  await scanLibraries(state.libraries.filter(item => item.type === 'movie' && (!selected || item.id === selected)));
}));
