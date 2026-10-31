import type { Job, Scan } from './types';

export function clock(seconds = 0): string {
  const value = Math.max(0, Math.floor(seconds));
  const hours = Math.floor(value / 3600);
  const minutes = Math.floor((value % 3600) / 60);
  const rest = String(value % 60).padStart(2, '0');
  return hours ? `${hours}:${String(minutes).padStart(2, '0')}:${rest}` : `${minutes}:${rest}`;
}
export const duration = (value: number) =>
  value > 0 ? `${Math.max(1, Math.round(value / 60))} 分钟` : '时长待识别';
export function bytes(value: number) {
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB'];
  let size = Math.max(0, value || 0),
    unit = 0;
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024;
    unit++;
  }
  return `${new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 1 }).format(size)} ${units[unit]}`;
}
export function resolution(item: { width: number; height: number }) {
  if (item.width >= 3800 || item.height >= 2100) return '4K';
  if (item.width >= 1900 || item.height >= 1000) return '1080p';
  if (item.width >= 1260 || item.height >= 700) return '720p';
  return item.width ? `${item.width} × ${item.height}` : '待识别';
}
export const activeJob = (status = '') => ['queued', 'running', 'scanning'].includes(status);
export function scanDescription(scan: Scan) {
  switch (scan.status) {
    case 'queued':
      return '扫描已排队，稍后开始';
    case 'scanning':
      return `正在扫描 · ${scan.processed} / ${scan.discovered}`;
    case 'completed':
      return `扫描完成 · ${scan.processed} 个视频${scan.errors ? ` · ${scan.errors} 项信息读取失败` : ''}`;
    case 'failed':
      return '扫描失败，请检查目录访问权限后重试';
    default:
      return '扫描已中断，可以重新扫描';
  }
}
const scrapeMessages: Record<string, string> = {
  ambiguous: '有多个相近候选，请手动匹配',
  metadata_conflict: '资料已更新或由 Agent 管理，请读取最新版本后再操作',
  not_found: '未找到匹配资料，请更换关键词或指定条目 ID',
  invalid_credential: '资料来源拒绝访问，请检查元数据设置',
  rate_limited: '资料来源请求过于频繁，请稍后重试',
  provider_unavailable: '无法读取资料来源，请检查网络后重试',
  poster_unavailable: '海报未下载成功，保留现有封面',
  no_chinese_title: '暂无中文译名，已保留原名',
  interrupted: '任务已中断，可以重新刮削',
  storage_error: '资料保存失败，请稍后重试',
};
export const scrapeMessage = (value = '') =>
  value
    .split(',')
    .filter(Boolean)
    .map((code) => scrapeMessages[code] || '刮削未完成，请稍后重试')
    .join('；');
export const stateNames: Record<string, string> = {
  matched: '已识别',
  unmatched: '未识别',
  needs_review: '待确认',
  failed: '失败',
  warning: '有提示',
  queued: '排队中',
  running: '运行中',
  completed: '已完成',
  interrupted: '已中断',
  pending: '待处理',
  skipped: '已跳过',
  not_processed: '未执行',
};
export const time = (value: number) =>
  value ? new Date(value * 1000).toLocaleString('zh-CN', { hour12: false }) : '—';
export function jobDescription(job: Job) {
  return `${stateNames[job.status] || job.status} · ${job.processed}/${job.total} · 匹配 ${job.matched} · 待确认 ${job.review} · 失败 ${job.errors}${job.message ? ' · ' + scrapeMessage(job.message) : ''}`;
}
