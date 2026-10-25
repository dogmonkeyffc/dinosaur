'use strict';

const seriesScraping = { timer: null, request: 0, signature: '', active: false };
function stopSeriesScraping() {
  clearTimeout(seriesScraping.timer);
  seriesScraping.request++;
  seriesScraping.signature = '';
  seriesScraping.active = false;
}
async function loadSeriesScrapeJobs() {
  if (state.user?.role !== 'admin') return;
  clearTimeout(seriesScraping.timer);
  const request = ++seriesScraping.request;
  const userID = state.user.id;
  try {
    const { jobs } = await api('/api/admin/tv/metadata/jobs');
    if (request !== seriesScraping.request || state.user?.id !== userID) return;
    const job = jobs[0];
    seriesScraping.active = jobs.some(item => ['queued', 'running'].includes(item.status));
    $('scrape-series').disabled = seriesScraping.active;
    $('series-scrape-status').hidden = !job;
    $('series-scrape-status').textContent = job ? scrapeJobDescription(job) : '';
    const signature = job ? `${job.id}:${job.status}:${job.matched}:${job.review}:${job.errors}` : '';
    const changed = seriesScraping.signature && signature !== seriesScraping.signature;
    seriesScraping.signature = signature;
    if (changed && state.section === 'scraping') loadScrapeAdmin({ quiet: true });
    if (changed && !seriesScraping.active) {
      if (state.section === 'series') loadSeries();
      if (state.section === 'series-detail') openSeries(series.id, { replace: true });
    }
    if (seriesScraping.active) seriesScraping.timer = setTimeout(loadSeriesScrapeJobs, 1500);
  } catch (error) {
    if (request === seriesScraping.request && state.user?.id === userID) handleError(error);
  }
}
async function startSeriesScraping(body, button) {
  const userID = state.user?.id;
  await busy(button, '正在开始刮削…', async () => {
    try {
      await api('/api/admin/tv/metadata/jobs', { method: 'POST', body });
      if (state.user?.id !== userID) return;
      seriesScraping.signature = 'pending';
      notify('剧集刮削已开始，每部剧只获取一次资料。');
      await loadSeriesScrapeJobs();
    } catch (error) { if (state.user?.id === userID) handleError(error); }
  });
  if (button.id === 'scrape-series') button.disabled = seriesScraping.active;
}
$('scrape-series').addEventListener('click', event => startSeriesScraping({ library_id: Number($('series-library-filter').value) }, event.currentTarget));
