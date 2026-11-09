import { createDOMRenderer, FluentProvider, RendererProvider } from '@fluentui/react-components';
import { useSyncExternalStore, type ReactNode } from 'react';
import { darkTheme, lightTheme } from './components/theme';
import { styleNonce } from './lib/styleNonce';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { AuthProvider } from './auth/AuthProvider';
import { ToastProvider } from './components/Toast';
import { App } from './App';
import './styles/app.css';
import './styles/metadata-settings.css';
import './styles/system-settings.css';
import './styles/devices.css';
import './styles/series.css';
import './styles/scrape-admin.css';
import './styles/movie-metadata.css';
import './styles/movies.css';
import './styles/movie-detail.css';
import './styles/player.css';
import './styles/watch-progress.css';
import './styles/layout.css';
import './styles/player-page.css';
import './styles/player-options.css';
import './styles/mobile-layout.css';
import './styles/agent-keys.css';

const systemAppearance = window.matchMedia('(prefers-color-scheme: dark)');

function subscribeAppearance(update: () => void) {
  systemAppearance.addEventListener('change', update);
  return () => systemAppearance.removeEventListener('change', update);
}

function Appearance({ children }: { children: ReactNode }) {
  const dark = useSyncExternalStore(subscribeAppearance, () => systemAppearance.matches);
  return (
    <FluentProvider theme={dark ? darkTheme : lightTheme} className="dinosaur-ui">
      {children}
    </FluentProvider>
  );
}

const renderer = createDOMRenderer(document, {
  styleElementAttributes: styleNonce ? { nonce: styleNonce } : undefined,
});

createRoot(document.getElementById('root')!).render(
  <RendererProvider renderer={renderer}>
    <Appearance>
      <BrowserRouter>
        <ToastProvider>
          <AuthProvider>
            <App />
          </AuthProvider>
        </ToastProvider>
      </BrowserRouter>
    </Appearance>
  </RendererProvider>,
);
