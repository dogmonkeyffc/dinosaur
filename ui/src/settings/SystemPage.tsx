import { ErrorMessage, Heading, Loading } from '../components/common';
import { bytes } from '../lib/format';
import { useResource } from '../lib/hooks';

interface Disk {
  mount_point: string;
  total_bytes: number;
  used_bytes: number;
  available_bytes: number;
}
export function SystemPage() {
  const { data, loading, error, reload } = useResource<{ disks: Disk[]; incomplete: boolean }>(
    '/api/admin/system/disks',
  );
  return (
    <section id="system-section">
      <Heading title="系统设置" />
      <div className="disk-section-heading">
        <h2>磁盘空间</h2>
        <button id="disk-refresh" className="button secondary" disabled={loading} onClick={reload}>
          {loading ? '读取中…' : '刷新'}
        </button>
      </div>
      <ErrorMessage error={error} />
      {loading && !data && <Loading />}
      {data?.incomplete && (
        <p className="muted" role="status">
          部分磁盘暂时无法读取，当前仅显示可读取的磁盘。
        </p>
      )}
      {!loading && !error && data && !data.disks.length && (
        <p className="muted" role="status">
          暂无可显示的已挂载磁盘。
        </p>
      )}
      <div id="disk-list" className="disk-grid" aria-label="磁盘空间" aria-busy={loading}>
        {!error &&
          data?.disks.map((disk) => {
            const capacity = disk.used_bytes + disk.available_bytes;
            const percent = capacity ? Math.min(100, (disk.used_bytes / capacity) * 100) : 0;
            return (
              <article key={disk.mount_point} className="disk-card">
                <div className="disk-heading">
                  <h3 title={disk.mount_point}>
                    {disk.mount_point === '/' ? '系统分区 /' : disk.mount_point}
                  </h3>
                  <span>{Math.ceil(percent)}%</span>
                </div>
                <progress
                  className="disk-progress"
                  value={percent}
                  max={100}
                  aria-label={`${disk.mount_point} 已用空间比例`}
                />
                <dl className="disk-capacity">
                  <div>
                    <dt>已用</dt>
                    <dd>{bytes(disk.used_bytes)}</dd>
                  </div>
                  <div>
                    <dt>可用</dt>
                    <dd>{bytes(disk.available_bytes)}</dd>
                  </div>
                </dl>
              </article>
            );
          })}
      </div>
    </section>
  );
}
