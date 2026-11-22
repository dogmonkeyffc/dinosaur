import { Button, Checkbox, FormControlLabel, Radio, RadioGroup, Stack } from '@mui/material';
import { useState } from 'react';
import { ErrorMessage, Loading, Modal } from '../components/common';
import { api } from '../lib/api';
import { useAction, useResource } from '../lib/hooks';
import type { Library } from '../lib/types';

interface AccessData {
  access: { restricted: boolean; user_ids: number[] };
  accounts: { id: number; username: string; disabled: boolean }[];
}

export function LibraryAccessDialog({ item, close }: { item: Library; close: () => void }) {
  const path = `/api/admin/libraries/${item.id}/access`;
  const resource = useResource<AccessData>(path);
  return (
    <Modal title={`给账号赋权 · ${item.name}`} close={close}>
      <ErrorMessage error={resource.error} />
      {resource.error && <Button onClick={resource.reload}>重试</Button>}
      {resource.loading && <Loading />}
      {resource.data && <LibraryAccessForm data={resource.data} path={path} close={close} />}
    </Modal>
  );
}

function LibraryAccessForm({ data, path, close }: { data: AccessData; path: string; close: () => void }) {
  const [restricted, setRestricted] = useState(data.access.restricted);
  const [selected, setSelected] = useState(data.access.user_ids);
  const action = useAction();
  return (
    <form onSubmit={(event) => {
      event.preventDefault();
      void action.run(async () => {
        await api(path, { method: 'PUT', body: { restricted, user_ids: selected } });
        close();
      });
    }}>
      <RadioGroup value={restricted ? 'restricted' : 'all'} onChange={(_, value) => setRestricted(value === 'restricted')} aria-label="可访问账号">
        <FormControlLabel value="all" control={<Radio />} label="所有账号" disabled={action.busy} />
        <FormControlLabel value="restricted" control={<Radio />} label="仅勾选的账号和管理员" disabled={action.busy} />
      </RadioGroup>
      {restricted && (
        <Stack sx={{ maxHeight: '40vh', overflow: 'auto', px: 1 }}>
          {data.accounts.map((account) => (
            <FormControlLabel key={account.id} label={account.username + (account.disabled ? '（已停用）' : '')}
              control={<Checkbox checked={selected.includes(account.id)} onChange={(_, checked) => {
                setSelected((ids) => checked ? [...ids, account.id] : ids.filter((id) => id !== account.id));
              }} disabled={action.busy} />} />
          ))}
          {!data.accounts.length && <p className="muted">暂无普通账号</p>}
        </Stack>
      )}
      <ErrorMessage error={action.error} />
      <div className="modal-actions">
        <Button onClick={close} disabled={action.busy}>取消</Button>
        <Button variant="contained" type="submit" disabled={action.busy}>保存权限</Button>
      </div>
    </form>
  );
}
