import { Button } from '@mui/material';
import { Navigate, Route, Routes } from 'react-router-dom';
import { lazy, type ReactNode } from 'react';
import { useAuth } from './auth/AuthProvider';
import { AuthPage } from './auth/AuthPage';
import { Brand } from './components/common';
import { Shell } from './layout/Shell';
import { LibraryProvider } from './libraries/LibraryProvider';
import { LibrariesPage } from './libraries/LibrariesPage';
import { HomePage } from './catalog/HomePage';
import { CatalogPage } from './catalog/CatalogPage';
import { MovieDetailPage, SeriesDetailPage } from './catalog/DetailPages';

const AccountPage = lazy(() =>
  import('./settings/AccountPage').then((page) => ({ default: page.AccountPage })),
);
const SystemPage = lazy(() => import('./settings/SystemPage').then((page) => ({ default: page.SystemPage })));
const MetadataPage = lazy(() =>
  import('./settings/MetadataPage').then((page) => ({ default: page.MetadataPage })),
);
const ScrapingPage = lazy(() =>
  import('./scraping/ScrapingPage').then((page) => ({ default: page.ScrapingPage })),
);
const PlayerPage = lazy(() => import('./playback/PlayerPage').then((page) => ({ default: page.PlayerPage })));
function Admin({ children }: { children: ReactNode }) {
  return useAuth().user?.role === 'admin' ? children : <Navigate to="/settings" replace />;
}

export function App() {
  const auth = useAuth();
  if (auth.status === 'loading' || auth.status === 'error')
    return (
      <div className="boot">
        <Brand />
        <p>{auth.error || '正在连接你的媒体空间…'}</p>
        {auth.error && (
          <Button variant="outlined" onClick={auth.boot}>
            重新连接
          </Button>
        )}
      </div>
    );
  if (auth.status !== 'ready') return <AuthPage />;
  return (
    <LibraryProvider key={auth.user!.id}>
      <Routes>
        <Route element={<Shell />}>
          <Route index element={<HomePage />} />
          <Route path="/settings" element={<LibrariesPage />} />
          <Route path="/settings/account" element={<AccountPage />} />
          <Route
            path="/settings/system"
            element={
              <Admin>
                <SystemPage />
              </Admin>
            }
          />
          <Route
            path="/settings/metadata"
            element={
              <Admin>
                <MetadataPage />
              </Admin>
            }
          />
          <Route
            path="/settings/scraping"
            element={
              <Admin>
                <ScrapingPage />
              </Admin>
            }
          />
          <Route path="/movies" element={<CatalogPage key="movie" kind="movie" />} />
          <Route path="/series" element={<CatalogPage key="tv" kind="tv" />} />
          <Route path="/movies/:id" element={<MovieDetailPage />} />
          <Route path="/series/:id" element={<SeriesDetailPage />} />
          <Route path="/watch/:id" element={<PlayerPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </LibraryProvider>
  );
}
