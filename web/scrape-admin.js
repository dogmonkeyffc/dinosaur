'use strict';

const scrapeAdmin = { tab: 'records', recordsPage: 1, jobsPage: 1, detailPage: 1, jobID: null, request: 0, detailRequest: 0, timer: null };
function stopScrapeAdmin() {
  scrapeAdmin.request++;
  scrapeAdmin.detailRequest++;
  clearTimeout(scrapeAdmin.timer);
}
function resetScrapeAdmin() {
  stopScrapeAdmin();
  scrapeAdmin.recordsPage = scrapeAdmin.jobsPage = scrapeAdmin.detailPage = 1;
  scrapeAdmin.jobID = null;
  $('scrape-record-filters').reset();
  $('scrape-history-filter').reset();
  for (const id of ['scrape-record-body', 'scrape-history-body', 'scrape-job-body', 'scrape-summary']) $(id).replaceChildren();
  $('scrape-job-detail').hidden = true;
}
function syncScrapeAdminLibraries() {
  const selected = $('scrape-admin-library').value;
  $('scrape-admin-library').innerHTML = '<option value="">全部媒体库</option>' + state.libraries.filter(item => item.type === 'movie').map(item => `<option value="${item.id}">${escapeHTML(item.name)}</option>`).join('');
  if (state.libraries.some(item => String(item.id) === selected && item.type === 'movie')) $('scrape-admin-library').value = selected;
}
function setScrapeAdminTab(tab) {
  scrapeAdmin.tab = tab;
  for (const name of ['records', 'jobs']) {
    $(`scrape-admin-${name}`).hidden = tab !== name;
    document.querySelector(`[data-admin-tab="${name}"]`).setAttribute('aria-selected', String(tab === name));
  }
  $('scrape-summary').hidden = tab !== 'records';
  loadScrapeAdmin();
}
async function loadScrapeAdmin({ quiet = false } = {}) {
  if (state.user?.role !== 'admin' || state.section !== 'scraping') return;
  const revision = ++scrapeAdmin.request;
  const userID = state.user.id;
  const active = () => revision === scrapeAdmin.request && state.user?.id === userID && state.section === 'scraping';
  if (!quiet) showError('scrape-admin-error');
  const records = scrapeAdmin.tab === 'records';
  const query = new URLSearchParams({ page: records ? scrapeAdmin.recordsPage : scrapeAdmin.jobsPage, page_size: records ? 20 : 10 });
  if (records) {
    if ($('scrape-admin-library').value) query.set('library_id', $('scrape-admin-library').value);
    if ($('scrape-admin-query').value.trim()) query.set('q', $('scrape-admin-query').value.trim());
    if ($('scrape-admin-state').value) query.set('status', $('scrape-admin-state').value);
  } else if ($('scrape-job-state').value) query.set('status', $('scrape-job-state').value);
  try {
    const result = await api(`/api/admin/metadata/${records ? 'records' : 'jobs'}?${query}`);
    if (!active()) return;
    const pages = Math.max(1, Math.ceil(result.total / result.page_size));
    if (result.page > pages) {
      scrapeAdmin[records ? 'recordsPage' : 'jobsPage'] = pages;
      return loadScrapeAdmin({ quiet });
    }
    if (records) renderScrapeRecords(result);
    else {
      renderScrapeHistory(result);
      if (scrapeAdmin.jobID) loadScrapeJobDetail();
    }
    showError('scrape-admin-error');
  } catch (error) {
    if (active()) handleError(error, 'scrape-admin-error');
  }
}
async function loadScrapeJobDetail() {
  const revision = ++scrapeAdmin.detailRequest;
  const userID = state.user?.id;
  try {
    const result = await api(`/api/admin/metadata/jobs/${scrapeAdmin.jobID}?page=${scrapeAdmin.detailPage}`);
    if (revision !== scrapeAdmin.detailRequest || state.user?.id !== userID || state.section !== 'scraping' || scrapeAdmin.tab !== 'jobs') return;
    renderScrapeJobDetail(result);
  } catch (error) {
    if (revision === scrapeAdmin.detailRequest && state.user?.id === userID && state.section === 'scraping' && scrapeAdmin.tab === 'jobs') handleError(error, 'scrape-admin-error');
  }
}
$('scrape-record-filters').addEventListener('submit', event => event.preventDefault());
$('scrape-admin-query').addEventListener('input', () => {
  clearTimeout(scrapeAdmin.timer);
  scrapeAdmin.timer = setTimeout(() => { scrapeAdmin.recordsPage = 1; loadScrapeAdmin(); }, 250);
});
for (const id of ['scrape-admin-library', 'scrape-admin-state']) $(id).addEventListener('change', () => { scrapeAdmin.recordsPage = 1; loadScrapeAdmin(); });
$('scrape-job-state').addEventListener('change', () => { scrapeAdmin.jobsPage = 1; loadScrapeAdmin(); });
$('scrape-admin-refresh').addEventListener('click', () => { loadScrapeAdmin(); loadScrapeJobs(); });
$('scrape-admin-start').addEventListener('click', event => startScraping({ library_id: scrapeAdmin.tab === 'records' ? Number($('scrape-admin-library').value) : 0 }, event.currentTarget));
$('scrape-admin-config').addEventListener('click', () => switchSection('metadata'));
$('scraping-section').addEventListener('click', async event => {
  const tab = event.target.closest('[data-admin-tab]');
  if (tab) setScrapeAdminTab(tab.dataset.adminTab);
  const stateButton = event.target.closest('[data-admin-state]');
  if (stateButton) {
    $('scrape-admin-state').value = stateButton.dataset.adminState;
    scrapeAdmin.recordsPage = 1;
    loadScrapeAdmin();
  }
  const job = event.target.closest('[data-admin-job]');
  if (job) {
    scrapeAdmin.jobID = Number(job.dataset.adminJob);
    scrapeAdmin.detailPage = 1;
    await loadScrapeJobDetail();
    $('scrape-job-detail').scrollIntoView({ block: 'start' });
  }
  const previous = event.target.closest('[data-admin-prev]');
  const next = event.target.closest('[data-admin-next]');
  if (previous || next) {
    const group = previous?.dataset.adminPrev || next.dataset.adminNext;
    scrapeAdmin[group === 'records' ? 'recordsPage' : group === 'jobs' ? 'jobsPage' : 'detailPage'] += previous ? -1 : 1;
    if (group === 'detail') loadScrapeJobDetail();
    else loadScrapeAdmin();
  }
  const retry = event.target.closest('[data-admin-retry]');
  if (retry) await startScraping({ media_id: Number(retry.dataset.adminRetry), refresh: retry.dataset.refresh === 'true' }, retry);
  const view = event.target.closest('[data-admin-view], [data-admin-match]');
  if (view) {
    await openMovie(Number(view.dataset.adminView || view.dataset.adminMatch));
    if (view.hasAttribute('data-admin-match')) {
      const editor = $('movie-detail').querySelector('.movie-matching');
      if (editor) { editor.open = true; editor.scrollIntoView({ block: 'start' }); }
    }
  }
});
