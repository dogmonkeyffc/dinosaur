import { Button, FormControlLabel, List, ListItem, Switch } from '@mui/material';
import { useState } from 'react';
import { ErrorMessage, Loading, Modal } from '../components/common';
import { useAction, useResource } from '../lib/hooks';
import { api } from '../lib/api';

interface LibrarySetting {
  id: number;
  name: string;
  record_history: boolean;
}

export function HistoryRecordingSettings({ close }: { close: () => void }) {
  const { data, loading, error, reload } = useResource<{ libraries: LibrarySetting[] }>(
    '/api/history/settings',
  );
  const action = useAction();
  const [saved, setSaved] = useState<Record<number, boolean>>({});
  function toggle(id: number, record_history: boolean) {
    void action.run(async () => {
      await api(`/api/history/settings/${id}`, { method: 'PUT', body: { record_history } });
      setSaved((values) => ({ ...values, [id]: record_history }));
    });
  }
  return (
    <Modal title="记录观看" close={close}>
      <ErrorMessage error={error || action.error} />
      {!!error && <Button onClick={reload}>重试</Button>}
      {loading && <Loading />}
      {!loading && !error && data?.libraries.length === 0 && <p>暂无媒体库</p>}
      <List>
        {data?.libraries.map((item) => (
          <ListItem key={item.id}>
            <FormControlLabel
              label={item.name}
              control={
                <Switch
                  checked={saved[item.id] ?? item.record_history}
                  disabled={loading || action.busy}
                  onChange={(_, enabled) => toggle(item.id, enabled)}
                />
              }
            />
          </ListItem>
        ))}
      </List>
    </Modal>
  );
}
