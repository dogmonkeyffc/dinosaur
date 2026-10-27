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
import './styles/detail-layout.css';
import './styles/player.css';
import './styles/watch-progress.css';
import './styles/layout.css';
import './styles/player-page.css';
import './styles/player-options.css';
import './styles/mobile-layout.css';

createRoot(document.getElementById('root')!).render(
  <BrowserRouter>
    <ToastProvider>
      <AuthProvider>
        <App />
      </AuthProvider>
    </ToastProvider>
  </BrowserRouter>,
);
