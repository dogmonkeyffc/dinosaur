import { SourceReference } from '../metadata/SourceReference';
import { Select } from '../components/Select';
import { useState } from 'react';
import { ErrorMessage, Loading, Pagination } from '../components/common';
import { useResource } from '../lib/hooks';
import { jobDescription, scrapeMessage, stateNames, time } from '../lib/format';
import type { Job, ScrapeRecord } from '../lib/types';
import { RecordActions } from './RecordActions';

interface JobPage {
  jobs: Job[];
  total: number;
  page_size: number;
}
interface Detail {
  job: Job;
  items: ScrapeRecord[];
  total: number;
  page_size: number;
}
export function Badge({ status }: { status: string }) {
  return <span className={`scrape-state scrape-state-${status}`}>{stateNames[status] || status}</span>;
}

export function JobsView() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState('');
  const [job, setJob] = useState<number>();
  const resource = useResource<JobPage>(
    `/api/admin/metadata/jobs?page=${page}&page_size=10${status ? '&status=' + status : ''}`,
    3000,
  );
  return (
    <>
      <form className="scrape-admin-filters" onSubmit={(event) => event.preventDefault()}>
        <label>
          任务状态
          <Select
            value={status}
            onValueChange={(value) => {
              setStatus(value);
              setPage(1);
            }}
          >
            <option value="">全部任务</option>
            {['queued', 'running', 'completed', 'failed', 'interrupted'].map((item) => (
              <option key={item} value={item}>
                {stateNames[item]}
              </option>
            ))}
          </Select>
        </label>
      </form>
      <ErrorMessage error={resource.error} />
      {resource.loading && !resource.data && <Loading />}
      <div className="scrape-table-wrap" role="region" aria-label="刮削任务历史" tabIndex={0}>
        <table className="scrape-table">
          <thead>
            <tr>
              <th>任务 / 范围</th>
              <th>状态 / 进度</th>
              <th>结果 / 提示</th>
              <th>开始 / 结束时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {resource.data?.jobs.map((item) => (
              <tr key={item.id}>
                <td>
                  <strong>#{item.id}</strong>
                  <small>
                    {item.media_id
                      ? `${item.media_type === 'tv' ? '剧集' : '影片'} #${item.media_id}`
                      : item.library_id
                        ? `媒体库 #${item.library_id}`
                        : item.media_type === 'tv'
                          ? '所有未识别剧集'
                          : '所有未识别影片'}
                  </small>
                </td>
                <td>
                  <Badge status={item.status} />
                  <small>
                    {item.processed} / {item.total}
                  </small>
                </td>
                <td>
                  匹配 {item.matched} · 待确认 {item.review}
                  <small>
                    失败 {item.errors} · 跳过 {item.skipped}
                  </small>
                  <p className="scrape-row-note">{scrapeMessage(item.message)}</p>
                </td>
                <td>
                  {time(item.created_at)}
                  <small>结束：{time(item.ended_at)}</small>
                </td>
                <td>
                  <button className="text-button" onClick={() => setJob(item.id)}>
                    查看结果
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {resource.data && !resource.data.total && <p className="scrape-admin-empty">还没有刮削任务。</p>}
      <Pagination page={page} total={resource.data?.total || 0} size={10} onChange={setPage} />
      {job && <JobDetail key={job} id={job} />}
    </>
  );
}

function JobDetail({ id }: { id: number }) {
  const [page, setPage] = useState(1);
  const resource = useResource<Detail>(`/api/admin/metadata/jobs/${id}?page=${page}`, 3000);
  return (
    <section className="scrape-job-detail">
      <h2>
        任务 #{id} · {resource.data ? stateNames[resource.data.job.status] : ''}
      </h2>
      <p className="scrape-status">{resource.data && jobDescription(resource.data.job)}</p>
      <ErrorMessage error={resource.error} />
      <div className="scrape-table-wrap" role="region" aria-label="任务逐片结果" tabIndex={0}>
        <table className="scrape-table">
          <thead>
            <tr>
              <th>媒体 / 文件</th>
              <th>当时状态 / 提示</th>
              <th>资料来源 / ID</th>
              <th>记录时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            {resource.data?.items.map((item, index) => (
              <tr key={index}>
                <td>
                  <strong>{item.title}</strong>
                  <small>
                    {item.filename} · {item.library_name}
                  </small>
                </td>
                <td>
                  <Badge status={item.status} />
                  <p className="scrape-row-note">{scrapeMessage(item.message)}</p>
                </td>
                <td>
                  <SourceReference item={item} kind={item.series_id ? 'tv' : 'movie'} />
                </td>
                <td>{time(item.updated_at)}</td>
                <td>
                  <RecordActions item={item} reload={resource.reload} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {resource.data && !resource.data.total && <p className="scrape-admin-empty">此任务暂无逐片记录。</p>}
      <Pagination
        page={page}
        total={resource.data?.total || 0}
        size={resource.data?.page_size || 20}
        onChange={setPage}
      />
    </section>
  );
}
