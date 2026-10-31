import { Link, useLocation, useNavigate } from 'react-router-dom';
import { ErrorMessage, Icon } from '../components/common';
import { useAction } from '../lib/hooks';
import { bytes, clock, duration, resolution } from '../lib/format';
import { saveProgress } from '../lib/progress';
import type { Movie } from '../lib/types';
import { Poster } from './MediaCard';

export function DetailHeader({ series = false }: { series?: boolean }) {
  const navigate = useNavigate();
  const location = useLocation();
  return (
    <header className="detail-page-header">
      <button
        className="icon-button detail-back"
        aria-label="返回上一页"
        onClick={() => navigate(location.state?.from || (series ? '/series' : '/movies'))}
      >
        <Icon name="arrow" />
      </button>
      <span>{series ? '剧集详情' : '电影详情'}</span>
    </header>
  );
}
export function Artwork({ url, title, series = false }: { url?: string; title: string; series?: boolean }) {
  return (
    <aside className="detail-artwork">
      <div className="detail-poster-frame">
        <span className="detail-poster-placeholder">
          <Icon name={series ? 'tv' : 'film'} />
        </span>
        <Poster url={url} title={title} className="detail-poster" />
      </div>
    </aside>
  );
}
export function WatchedButton({ item, reload }: { item: Movie; reload: () => void }) {
  const action = useAction();
  return (
    <>
      <button
        className="text-button"
        disabled={action.busy}
        onClick={() =>
          action.run(async () => {
            await saveProgress(item.id, { watched: !item.watched });
            reload();
          })
        }
      >
        {item.watched ? '标记为未看' : '标记为已看'}
      </button>
      <ErrorMessage error={action.error} />
    </>
  );
}
export function PlayActions({ item, reload }: { item: Movie; reload: () => void }) {
  const resume = item.position > 0 && !item.watched;
  return (
    <div className="movie-play-actions">
      <Link className="button primary" to={`/watch/${item.id}`} state={{ kind: item.type }}>
        {resume ? `继续观看 · ${clock(item.position)}` : '播放电影'}
      </Link>
      {resume && (
        <Link className="button secondary" to={`/watch/${item.id}?beginning=1`} state={{ kind: item.type }}>
          从头播放
        </Link>
      )}
      <WatchedButton item={item} reload={reload} />
    </div>
  );
}
export function TechnicalInfo({ item }: { item: Movie }) {
  return (
    <details className="movie-technical">
      <summary>媒体信息</summary>
      <dl className="movie-facts">
        <div>
          <dt>视频编码</dt>
          <dd>{item.video_codec?.toUpperCase() || '待识别'}</dd>
        </div>
        <div>
          <dt>分辨率</dt>
          <dd>{item.width && item.height ? `${item.width} × ${item.height}` : '待识别'}</dd>
        </div>
        <div>
          <dt>文件大小</dt>
          <dd>{bytes(item.size)}</dd>
        </div>
      </dl>
      {item.tracks?.length > 0 && (
        <>
          <h3>音轨与字幕</h3>
          <ul className="movie-tracks">
            {item.tracks.map((track) => (
              <li key={track.index}>
                <strong>
                  {track.type === 'audio' ? '音轨' : '字幕'} {track.index}
                </strong>
                <span>
                  {[track.language, track.codec, track.channels ? `${track.channels} 声道` : '', track.title]
                    .filter(Boolean)
                    .join(' · ')}
                </span>
              </li>
            ))}
          </ul>
        </>
      )}
      <div className="movie-file">
        <h3>文件信息</h3>
        <p>{item.filename}</p>
        <p>容器：{item.format || '待识别'}</p>
      </div>
    </details>
  );
}
export function EpisodeCard({ item, reload }: { item: Movie; reload: () => void }) {
  const resume = item.position > 0 && !item.watched;
  return (
    <article className="series-episode">
      <div>
        <h2>{item.title}</h2>
        <p className="muted">
          {duration(item.duration)} · {resolution(item)}
          {item.watched ? ' · 已看' : item.position > 0 ? ` · 已看 ${clock(item.position)}` : ''}
        </p>
        {item.position > 0 && item.duration > 0 && (
          <progress value={item.position} max={item.duration} aria-label="观看进度" />
        )}
        <TechnicalInfo item={item} />
        {item.probe_error && <p className="field-help">视频信息读取失败</p>}
      </div>
      <div className="series-episode-actions">
        <Link className="button secondary" to={`/watch/${item.id}`} state={{ kind: 'tv' }}>
          {resume ? '继续观看' : '播放'}
        </Link>
        {resume && (
          <Link className="text-button" to={`/watch/${item.id}?beginning=1`} state={{ kind: 'tv' }}>
            从头播放
          </Link>
        )}
        <WatchedButton item={item} reload={reload} />
      </div>
    </article>
  );
}
