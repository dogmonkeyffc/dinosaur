'use strict';

const scrapeStateNames = {
  matched: '已识别', unmatched: '未识别', needs_review: '待确认', failed: '失败', warning: '已识别但有提示',
  queued: '排队中', running: '运行中', completed: '已完成', interrupted: '已中断',
  pending: '待处理', skipped: '已跳过', not_processed: '未执行',
};
const scrapeTime = value => value ? new Date(value * 1000).toLocaleString('zh-CN', { hour12: false }) : '—';
const scrapeStateBadge = status => `<span class="scrape-state scrape-state-${escapeHTML(status)}">${escapeHTML(scrapeStateNames[status] || status)}</span>`;
function scrapeAdminActions(item) {
  if (!item.media_id) return '<span class="muted">影片已移除</span>';
  return `<div class="scrape-row-actions"><button class="text-button" data-admin-view="${item.media_id}">详情</button><button class="text-button" data-admin-match="${item.media_id}">匹配更正</button><button class="text-button" data-admin-retry="${item.media_id}" data-refresh="${item.tmdb_id > 0}" ${scraping.active ? 'disabled' : ''}>${item.tmdb_id > 0 ? '刷新资料' : '重试识别'}</button></div>`;
}
function renderScrapeSummary(summary) {
  const labels = { '': '全部影片', matched: '已识别', unmatched: '未识别', needs_review: '待确认', failed: '失败', warning: '有提示' };
  $('scrape-summary').innerHTML = Object.entries(labels).map(([status, label]) => `
    <button class="scrape-summary-card" data-admin-state="${status}" aria-pressed="${$('scrape-admin-state').value === status}"><strong>${summary[status || 'all'] || 0}</strong><span>${label}</span></button>`).join('');
}
function renderScrapeRecords(page) {
  renderScrapeSummary(page.summary);
  $('scrape-record-body').innerHTML = page.items.map(item => `<tr>
    <td><strong>${escapeHTML(item.title)}</strong><small>${escapeHTML(item.filename)}</small></td>
    <td>${escapeHTML(item.library_name)}</td>
    <td>${scrapeStateBadge(item.status)}${item.message ? `<p class="scrape-row-note">${escapeHTML(scrapeMessage(item.message))}</p>` : ''}${candidatePreview(item.media_id, item.candidates || [], item.message, true)}</td>
    <td>${item.tmdb_id || '—'}<small>${escapeHTML(item.language)}</small></td>
    <td>${scrapeTime(item.updated_at)}</td><td>${scrapeAdminActions(item)}</td>
  </tr>`).join('');
  $('scrape-record-empty').hidden = page.total > 0;
  updateScrapePagination('records', page);
}
function renderScrapeHistory(page) {
  $('scrape-history-body').innerHTML = page.jobs.map(job => `<tr>
    <td><strong>#${job.id}</strong><small>${job.media_id ? `影片 #${job.media_id}` : job.library_id ? `媒体库 #${job.library_id}` : '所有未识别影片'}</small></td>
    <td>${scrapeStateBadge(job.status)}<small>${job.processed} / ${job.total}</small></td>
    <td>匹配 ${job.matched} · 待确认 ${job.review}<small>失败 ${job.errors} · 跳过 ${job.skipped}</small>${job.message ? `<p class="scrape-row-note">${escapeHTML(scrapeMessage(job.message))}</p>` : ''}</td>
    <td>${scrapeTime(job.created_at)}<small>结束：${scrapeTime(job.ended_at)}</small></td>
    <td><button class="text-button" data-admin-job="${job.id}" aria-label="查看任务 ${job.id}">查看结果</button></td>
  </tr>`).join('');
  $('scrape-history-empty').hidden = page.total > 0;
  updateScrapePagination('jobs', page);
}
function renderScrapeJobDetail(detail) {
  $('scrape-job-detail').hidden = false;
  $('scrape-job-title').textContent = `任务 #${detail.job.id} · ${scrapeStateNames[detail.job.status] || detail.job.status}`;
  $('scrape-job-description').textContent = scrapeJobDescription(detail.job);
  $('scrape-job-body').innerHTML = detail.items.map(item => `<tr>
    <td><strong>${escapeHTML(item.title)}</strong><small>${escapeHTML(item.filename)} · ${escapeHTML(item.library_name)}</small></td>
    <td>${scrapeStateBadge(item.status)}${item.message ? `<p class="scrape-row-note">${escapeHTML(scrapeMessage(item.message))}</p>` : ''}</td>
    <td>${item.tmdb_id || '—'}</td><td>${scrapeTime(item.updated_at)}</td>
    <td>${scrapeAdminActions(item)}</td>
  </tr>`).join('');
  $('scrape-job-empty').hidden = detail.total > 0;
  $('scrape-job-empty').textContent = detail.job.total ? '此任务尚无逐片记录；升级前的任务仅保留汇总，可在“影片状态”中查看当前结果。' : '此次任务没有需要识别的影片。';
  updateScrapePagination('detail', detail);
}
function updateScrapePagination(group, page) {
  const pages = Math.max(1, Math.ceil(page.total / page.page_size));
  $(`scrape-${group}-page`).textContent = `共 ${page.total} 条 · ${page.page} / ${pages}`;
  document.querySelector(`[data-admin-prev="${group}"]`).disabled = page.page <= 1;
  document.querySelector(`[data-admin-next="${group}"]`).disabled = page.page >= pages;
}
