import { useEffect, useRef } from 'react';
import { useAuth } from '../auth/AuthProvider';
import { api } from '../lib/api';
import { activeJob } from '../lib/format';
import { useAction, useResource } from '../lib/hooks';
import type { Job, MediaKind } from '../lib/types';

export function useJobs(kind: MediaKind, onChange?: () => void) {
  const { user } = useAuth();
  const endpoint = kind === 'tv' ? '/api/admin/tv/metadata/jobs' : '/api/admin/metadata/jobs';
  const resource = useResource<{ jobs: Job[] }>(
    user?.role === 'admin' ? endpoint + (kind === 'movie' ? '?type=movie' : '') : null,
    3000,
  );
  const action = useAction();
  const latest = resource.data?.jobs[0];
  const running = resource.data?.jobs.some((job) => activeJob(job.status)) || false;
  const signature = latest
    ? `${latest.id}:${latest.status}:${latest.processed}:${latest.matched}:${latest.review}:${latest.errors}`
    : '';
  const previous = useRef('');
  useEffect(() => {
    if (previous.current && previous.current !== signature) onChange?.();
    previous.current = signature;
  }, [signature, onChange]);
  function start(body: Record<string, unknown> = {}) {
    return action.run(async () => {
      await api(endpoint, { method: 'POST', body });
      previous.current = 'pending';
      resource.reload();
    });
  }
  return { ...resource, ...action, latest, running, start };
}
