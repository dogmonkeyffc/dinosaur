import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../auth/AuthProvider';
import { ErrorMessage, Heading, Icon, Loading } from '../components/common';
import { useAction } from '../lib/hooks';
import { activeJob, scanDescription } from '../lib/format';
import type { Library } from '../lib/types';
import { useLibraries } from './LibraryProvider';
import { DeleteLibrary, LibraryDialog } from './LibraryDialog';

export function LibrariesPage() {
  const { user } = useAuth();
  const { libraries, loading, error, scan, reload } = useLibraries();
  const admin = user?.role === 'admin';
  const action = useAction();
  const [filter, setFilter] = useState('all');
  const [query, setQuery] = useState('');
  const [editing, setEditing] = useState<Library | null | undefined>();
  const [deleting, setDeleting] = useState<Library>();
  const items = libraries.filter(
    (item) =>
      (filter === 'all' || item.type === filter) &&
      item.name.toLocaleLowerCase().includes(query.trim().toLocaleLowerCase()),
  );
  return (
    <section id="libraries-section">
      <Heading title="媒体库">
        {admin && (
          <button className="button primary" onClick={() => setEditing(null)}>
            <Icon name="plus" />
            添加媒体库
          </button>
        )}
      </Heading>
      <div className="list-toolbar">
        <div className="filter-tabs" role="group" aria-label="媒体类型">
          {[
            ['all', '全部'],
            ['movie', '电影'],
            ['tv', '电视剧'],
          ].map(([value, label]) => (
            <button
              key={value}
              className={`filter-tab${filter === value ? ' active' : ''}`}
              aria-pressed={filter === value}
              onClick={() => setFilter(value)}
            >
              {label}
              {value === 'all' && <span>{libraries.length}</span>}
            </button>
          ))}
        </div>
        <label className="search-field">
          <Icon name="search" />
          <input
            type="search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="搜索媒体库"
            aria-label="搜索媒体库"
          />
        </label>
      </div>
      <ErrorMessage error={error || action.error} />
      {error && (
        <button className="text-button" onClick={reload}>
          重新读取
        </button>
      )}
      {loading && !libraries.length && <Loading />}
      <div className="library-grid" aria-live="polite">
        {items.map((item) => {
          const scanning = activeJob(item.scan?.status);
          return (
            <article key={item.id} className="library-card">
              <div className="card-top">
                <span className={`library-type-icon ${item.type === 'tv' ? 'tv' : ''}`}>
                  <Icon name={item.type === 'tv' ? 'tv' : 'film'} />
                </span>
                <span className="type-tag">{item.type === 'tv' ? '电视剧' : '电影'}</span>
              </div>
              <h2>{item.name}</h2>
              {item.root_path && (
                <p className="card-path" title={item.root_path}>
                  {item.root_path}
                </p>
              )}
              <div className="card-bottom">
                <span className={`card-status ${admin && !item.available ? 'unavailable' : ''}`}>
                  <span className="status-dot" />
                  {admin ? (item.available ? '目录可访问' : '目录暂不可用') : '已配置'}
                </span>
                {admin && (
                  <div className="card-actions">
                    <button
                      className="icon-button"
                      disabled={scanning}
                      onClick={() => setEditing(item)}
                      aria-label={`编辑 ${item.name}`}
                    >
                      <Icon name="edit" />
                    </button>
                    <button
                      className="icon-button"
                      disabled={scanning}
                      onClick={() => setDeleting(item)}
                      aria-label={`移除 ${item.name}`}
                    >
                      <Icon name="trash" />
                    </button>
                  </div>
                )}
              </div>
              <div className="library-browse-row">
                <Link
                  className="text-button"
                  to={`/${item.type === 'tv' ? 'series' : 'movies'}?library_id=${item.id}`}
                >
                  {item.type === 'tv' ? '查看剧集' : '查看电影'} <span>{item.media_count || 0}</span>
                </Link>
                {admin && (
                  <button
                    className="text-button"
                    disabled={scanning || action.busy}
                    onClick={() => action.run(() => scan([item]))}
                  >
                    {scanning ? '正在扫描…' : '重新扫描'}
                  </button>
                )}
              </div>
              {admin && item.scan && <p className="scan-summary">{scanDescription(item.scan)}</p>}
            </article>
          );
        })}
      </div>
      {!loading && !items.length && (
        <div className="empty-state">
          <h2>{libraries.length ? '没有找到匹配的媒体库' : '暂无媒体库'}</h2>
          {admin && !libraries.length && (
            <button className="button secondary" onClick={() => setEditing(null)}>
              添加媒体库
            </button>
          )}
        </div>
      )}
      {editing !== undefined && (
        <LibraryDialog item={editing || undefined} close={() => setEditing(undefined)} />
      )}
      {deleting && <DeleteLibrary item={deleting} close={() => setDeleting(undefined)} />}
    </section>
  );
}
