import createCache from '@emotion/cache';
import { CacheProvider } from '@emotion/react';
import { CssBaseline, ThemeProvider } from '@mui/material';
import { useSyncExternalStore, type ReactNode } from 'react';
import { darkTheme, lightTheme } from './components/theme';
import { styleNonce } from './lib/styleNonce';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { AuthProvider } from './auth/AuthProvider';
import { ToastProvider } from './components/Toast';
import { App } from './App';
import './styles/layout.css';
import './styles/media.css';
import './styles/management.css';
import './styles/player.css';
import './styles/player-controls.css';
import './styles/auth.css';

const systemAppearance = window.matchMedia('(prefers-color-scheme: dark)');

function subscribeAppearance(update: () => void) {
  systemAppearance.addEventListener('change', update);
  return () => systemAppearance.removeEventListener('change', update);
}

function Appearance({ children }: { children: ReactNode }) {
  const dark = useSyncExternalStore(subscribeAppearance, () => systemAppearance.matches);
  return (
    <ThemeProvider theme={dark ? darkTheme : lightTheme}>
      <CssBaseline enableColorScheme />
      {children}
    </ThemeProvider>
  );
}

const cache = createCache({ key: 'mui', nonce: styleNonce });

createRoot(document.getElementById('root')!).render(
  <CacheProvider value={cache}>
    <Appearance>
      <BrowserRouter>
        <ToastProvider>
          <AuthProvider>
            <App />
          </AuthProvider>
        </ToastProvider>
      </BrowserRouter>
    </Appearance>
  </CacheProvider>,
);
