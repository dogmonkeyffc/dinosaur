import { useEffect, useRef, useState } from 'react';
import { useToast } from '../components/Toast';
import { api } from '../lib/api';
import { useAction, useResource } from '../lib/hooks';
import { scrapeMessage } from '../lib/format';
import type { Job, MediaKind } from '../lib/types';

export function useManualMatch(kind: MediaKind, mediaID: number, reload: () => void) {
  const action = useAction();
  const toast = useToast();
  const [job, setJob] = useState<Job | null>(null);
  const [failure, setFailure] = useState('');
  const submitting = useRef(false);
  const result = useResource<{ job: Job }>(job ? `/api/admin/metadata/jobs/${job.id}` : null, 1000);
  useEffect(() => {
    const current = result.data?.job;
    if (!current || current.id !== job?.id || ['queued', 'running'].includes(current.status)) return;
    setJob(null);
    if (current.status === 'completed' && current.matched > 0 && !current.errors) toast('已确认匹配');
    else setFailure(scrapeMessage(current.message) || '匹配失败，请重试。');
    reload();
  }, [result.data, job?.id, reload, toast]);
  async function select(id: number, provider = 'tmdb') {
    if (submitting.current || job) return;
    submitting.current = true;
    setFailure('');
    try {
      await action.run(async () => {
        const response = await api<{ job: Job }>(`/api/admin/${kind === 'tv' ? 'tv/' : ''}metadata/jobs`, {
          method: 'POST',
          body: {
            [kind === 'tv' ? 'series_id' : 'media_id']: mediaID,
            [provider === 'bangumi' ? 'bangumi_id' : 'tmdb_id']: id,
          },
        });
        setJob(response.job);
        toast('已提交，优先匹配');
      });
    } finally {
      submitting.current = false;
    }
  }
  return {
    select,
    busy: action.busy || !!job,
    notice: job ? (result.data?.job.status === 'running' ? '正在匹配…' : '已排队，优先处理') : '',
    error: action.error || failure || result.error,
  };
}
