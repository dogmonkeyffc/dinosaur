import { useEffect, useRef, useState } from 'react';
import { useNavigate, useOutletContext, useParams, useSearchParams } from 'react-router-dom';
import { ErrorMessage } from '../components/common';
import { useToast } from '../components/Toast';
import { useResource } from '../lib/hooks';
import { message } from '../lib/errors';
import type { PlaybackInfo } from '../lib/types';
import { PlaybackEngine } from './engine';

export function PlayerPage() {
  const { id } = useParams();
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const notify = useToast();
  const { setPlayingKind } = useOutletContext<{ setPlayingKind: (kind: string) => void }>();
  const root = useRef<HTMLDivElement>(null);
  const engine = useRef<PlaybackEngine | null>(null);
  const [status, setStatus] = useState('正在准备…');
  const [error, setError] = useState('');
  const resource = useResource<PlaybackInfo>(`/api/playback/${id}?subtitle=auto`);
  const info = resource.data;
  const beginning = params.get('beginning') === '1';
  useEffect(() => {
    if (!info || !root.current) return;
    setPlayingKind(info.item.series_id ? 'tv' : 'movie');
    document.title = `${info.item.series_id ? info.item.series_title + ' · ' : ''}${info.item.title} · Dinosaur`;
    setError('');
    try {
      const instance = new PlaybackEngine(
        root.current,
        info,
        {
          status: setStatus,
          error: setError,
          notify,
          back: () =>
            navigate(info.item.series_id ? `/series/${info.item.series_id}` : '/movies', { replace: true }),
          next: (next) => navigate(`/watch/${next}?beginning=1`, { replace: true, state: { kind: 'tv' } }),
        },
        beginning,
      );
      engine.current = instance;
      return () => {
        instance.destroy();
        engine.current = null;
      };
    } catch (reason) {
      setError(message(reason));
    }
  }, [info, beginning, navigate, notify, setPlayingKind]);
  return (
    <section id="player-section" className="player-page" aria-label="媒体播放器">
      <div className="player-stage">
        <div ref={root} id="player-art" className="artplayer-host" />
      </div>
      <span className="player-accessible" role="status">
        {status}
      </span>
      {(error || resource.error) && (
        <div className="player-error-overlay">
          <ErrorMessage error={error || resource.error} />
          <button
            className="text-button"
            onClick={() =>
              engine.current ? engine.current.changeSource({ mode: 'compatible' }) : resource.reload()
            }
          >
            重试兼容模式
          </button>
        </div>
      )}
    </section>
  );
}
