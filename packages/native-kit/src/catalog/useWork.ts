import { useEffect, useState } from 'react';
import type { Client } from '@dinosaur/client';
import type { WorkDetail } from '@dinosaur/contracts';

export function useWork(client: Client, id: number) {
  const [detail, setDetail] = useState<WorkDetail>();
  const [error, setError] = useState('');
  const [revision, setRevision] = useState(0);
  useEffect(() => {
    let active = true;
    setDetail(undefined);
    setError('');
    if (!Number.isSafeInteger(id) || id <= 0) {
      setError('无效的作品。');
      return;
    }
    client.work(id).then((value) => {
      if (active) setDetail(value);
    }).catch((reason) => {
      if (active) setError(reason instanceof Error ? reason.message : '无法读取作品。');
    });
    return () => { active = false; };
  }, [client, id, revision]);
  return { detail, error, reload: () => setRevision((value) => value + 1) };
}
