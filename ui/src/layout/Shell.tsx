import { useEffect, useRef, useState } from 'react';
import { Link, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../auth/AuthProvider';
import { Brand, ErrorMessage, Icon } from '../components/common';
import { useAction, useResource } from '../lib/hooks';
import { useLibraries } from '../libraries/LibraryProvider';
import { libraryPath } from './navigation';

const mobileQuery = '(max-width:720px), (max-width:1000px) and (max-height:500px) and (pointer:coarse)';
const managementGroups = [
  {
    label: '媒体管理',
    entries: [
      ['/settings', '媒体库设置', 'folder'],
      ['/settings/scraping', '刮削管理', 'film', 'admin'],
      ['/settings/metadata', '元数据设置', 'info', 'admin'],
    ],
  },
  {
    label: '系统与账户',
    entries: [
      ['/settings/system', '系统设置', 'disk', 'admin'],
      ['/settings/account', '账户安全', 'lock'],
    ],
  },
];

export function Shell() {
  const { user, logout } = useAuth();
  const { libraries } = useLibraries();
  const location = useLocation();
  const [menu, setMenu] = useState(false);
  const [activeLibrary, setActiveLibrary] = useState<number | null>(null);
  const sidebar = useRef<HTMLElement>(null);
  const toggle = useRef<HTMLButtonElement>(null);
  const content = useRef<HTMLElement>(null);
  const action = useAction();
  const health = useResource<{ status: string }>('/health/ready', 30000);
  const playing = location.pathname.startsWith('/watch/');
  const detail = /^\/(movies|series)\/\d+$/.test(location.pathname);
  const navigationGroups = [
    {
      label: '媒体库',
      entries: libraries.map((library) => [
        libraryPath(library),
        library.name,
        library.type === 'movie' ? 'film' : 'tv',
      ]),
    },
    ...managementGroups,
  ].filter((group) => group.entries.length > 0);
  const entries = navigationGroups.flatMap((group) => group.entries);
  const libraryID =
    playing || detail ? activeLibrary : Number(new URLSearchParams(location.search).get('library_id'));
  const library = libraries.find((item) => item.id === libraryID);
  const catalog = location.pathname === '/movies' || location.pathname === '/series';
  const selected = library && (playing || detail || catalog) ? libraryPath(library) : location.pathname;
  const pageTitle = entries.find(([path]) => path === selected)?.[1] || '媒体空间';
  useEffect(() => {
    setMenu(false);
    content.current?.scrollTo(0, 0);
    if (!playing && !detail) document.title = `${pageTitle} · Dinosaur`;
  }, [location.pathname, selected, pageTitle]);
  useEffect(() => {
    const media = window.matchMedia(mobileQuery);
    const resize = () => setMenu(false);
    media.addEventListener('change', resize);
    return () => media.removeEventListener('change', resize);
  }, []);
  useEffect(() => {
    if (!menu) return;
    const key = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault();
        event.stopPropagation();
        setMenu(false);
        toggle.current?.focus();
      }
    };
    const outside = (event: PointerEvent) => {
      if (!sidebar.current?.contains(event.target as Node)) setMenu(false);
    };
    document.addEventListener('keydown', key, true);
    document.addEventListener('pointerdown', outside);
    return () => {
      document.removeEventListener('keydown', key, true);
      document.removeEventListener('pointerdown', outside);
    };
  }, [menu]);
  function close() {
    setMenu(false);
    if (menu) toggle.current?.focus();
  }
  return (
    <div id="app" className="app-shell">
      <aside ref={sidebar} className={`sidebar${menu ? ' menu-open' : ''}`}>
        <Brand />
        <button
          ref={toggle}
          id="mobile-menu-toggle"
          className="mobile-menu-toggle"
          type="button"
          aria-expanded={menu}
          aria-controls="workspace-navigation sidebar-account"
          aria-label={menu ? '收起导航菜单' : '展开导航菜单'}
          onClick={() => setMenu(!menu)}
        >
          <Icon name={menu ? 'close' : 'menu'} />
          <span>{menu ? '收起' : '菜单'}</span>
        </button>
        <nav id="workspace-navigation" aria-label="媒体空间">
          {navigationGroups.map((group) => (
            <div key={group.label} className="nav-group" role="group" aria-label={group.label}>
              {group.entries
                .filter(([, , , role]) => !role || user?.role === role)
                .map(([path, label, icon]) => (
                  <Link
                    key={path}
                    to={path}
                    aria-label={label}
                    title={label}
                    onClick={close}
                    className={`nav-item${selected === path ? ' active' : ''}`}
                    aria-current={selected === path ? 'page' : undefined}
                  >
                    <Icon name={icon} />
                    {label}
                    {path === '/settings' && <span className="nav-count">{libraries.length}</span>}
                  </Link>
                ))}
            </div>
          ))}
        </nav>
        <div className="sidebar-bottom">
          <div className="server-health">
            <span className={`status-dot ${health.error ? 'offline' : health.data ? 'online' : ''}`} />
            <div>
              <strong>{health.error ? '服务器连接异常' : health.data ? '服务器运行正常' : '正在连接'}</strong>
            </div>
          </div>
        </div>
        <div id="sidebar-account" className="sidebar-account user-area" role="group" aria-label="当前用户">
          <span className="avatar">{user?.username.slice(0, 1).toUpperCase()}</span>
          <div className="user-details">
            <strong>{user?.username}</strong>
            <span>{user?.role === 'admin' ? '管理员' : '普通用户'}</span>
          </div>
          <button
            id="logout"
            className="icon-button logout-button"
            aria-label="退出登录"
            title="退出登录"
            disabled={action.busy}
            onClick={() => action.run(logout)}
          >
            <Icon name="logout" />
          </button>
        </div>
      </aside>
      <div
        id="content"
        className={`workspace content${playing ? ' is-playing' : detail ? ' is-detail' : ''}`}
      >
        <main ref={content} id="content-main" className="main-content">
          <ErrorMessage error={action.error} />
          <Outlet context={{ setActiveLibrary }} />
        </main>
      </div>
    </div>
  );
}
