'use strict';

const scraping = { timer: null, request: 0, signature: '', active: false };
const scrapeMessages = {
  ambiguous: '有多个相近候选，请手动匹配', not_found: '未找到匹配影片，请更换关键词或填写 TMDB ID',
  invalid_credential: 'TMDB 凭据无效，请检查元数据设置', rate_limited: 'TMDB 请求过于频繁，请稍后重试',
  provider_unavailable: '无法读取 TMDB 资料，请检查网络后重试', poster_unavailable: '海报未下载成功，保留现有封面',
  no_chinese_title: '暂无中文译名，已保留原名', interrupted: '任务已中断，可以重新刮削', storage_error: '资料保存失败，请稍后重试',
};
const scrapeMessage = value => String(value || '').split(',').filter(Boolean).map(code => scrapeMessages[code] || '刮削未完成，请稍后重试').join('；');
function stopScrapePolling() {
  clearTimeout(scraping.timer);
  scraping.request++;
  scraping.signature = '';
  scraping.active = false;
}
function scrapeJobDescription(job) {
  const stateLabel = { queued: '已排队', running: '正在刮削', completed: '刮削完成', failed: '刮削失败', interrupted: '刮削已中断' }[job.status] || job.status;
  return `${stateLabel} · ${job.processed}/${job.total} · 匹配 ${job.matched} · 待确认 ${job.review} · 失败 ${job.errors}${job.message ? ' · ' + scrapeMessage(job.message) : ''}`;
}
async function loadScrapeJobs() {
  if (state.user?.role !== 'admin') return;
  clearTimeout(scraping.timer);
  const request = ++scraping.request;
  const userID = state.user.id;
  try {
    const { jobs } = await api('/api/admin/metadata/jobs');
    if (request !== scraping.request || state.user?.id !== userID) return;
    const job = jobs[0];
    const active = jobs.some(item => ['queued', 'running'].includes(item.status));
    scraping.active = active;
    for (const id of ['movie-scrape-status', 'settings-scrape-status']) {
      $(id).hidden = !job;
      $(id).textContent = job ? scrapeJobDescription(job) : '';
    }
    for (const id of ['scrape-movies', 'scrape-all']) $(id).disabled = active;
    const signature = job ? `${job.id}:${job.status}:${job.matched}:${job.review}:${job.errors}` : '';
    if (signature !== scraping.signature) {
      const hadResult = Boolean(scraping.signature);
      scraping.signature = signature;
      if (state.section === 'movies') loadMovies();
      if (hadResult && job && !active && $('movie-dialog').open && movieDetailID) openMovie(movieDetailID);
    }
    if (state.section === 'scraping') loadScrapeAdmin({ quiet: true });
    $('scrape-admin-start').disabled = active;
    if (active) scraping.timer = setTimeout(loadScrapeJobs, 1500);
  } catch (error) {
    if (request === scraping.request && state.user?.id === userID) handleError(error);
  }
}
async function startScraping(body, button) {
  const userID = state.user?.id;
  const start = async () => {
    try {
      await api('/api/admin/metadata/jobs', { method: 'POST', body });
      if (state.user?.id !== userID) return;
      notify('刮削任务已开始，已识别的影片会直接复用资料。');
      await loadScrapeJobs();
    } catch (error) {
      if (state.user?.id === userID) handleError(error);
    }
  };
  if (button) await busy(button, '正在开始刮削…', start);
  else await start();
  if (button && ['scrape-movies', 'scrape-all'].includes(button.id)) button.disabled = scraping.active;
}
$('scrape-movies').addEventListener('click', event => startScraping({ library_id: Number($('movie-library-filter').value) }, event.currentTarget));
$('scrape-all').addEventListener('click', event => startScraping({}, event.currentTarget));
