import { api, currentSession } from './api';

let writes: Promise<void> = Promise.resolve();
export const flushProgress = () => writes.catch(() => {});

export function saveProgress(id: number, body: { position?: number; watched?: boolean }, keepalive = false) {
  const identity = currentSession();
  writes = writes
    .catch(() => {})
    .then(async () => {
      if (!identity.session || currentSession().epoch !== identity.epoch) return;
      await api(`/api/playback/${id}/progress`, { method: 'PUT', body, keepalive });
    });
  return writes;
}
