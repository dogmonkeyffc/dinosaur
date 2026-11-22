import { Button, List, ListItemButton, ListItemIcon, ListItemText } from '@mui/material';
import { useState } from 'react';
import { ErrorMessage, Icon, Loading, Modal } from '../components/common';
import { useAction, useResource } from '../lib/hooks';
import { api } from '../lib/api';
import './directory-picker.css';

interface Folders {
  path: string;
  parent: string;
  folders: { name: string; path: string }[];
  next_after: string;
}

function directoryURL(path: string, after = '') {
  return '/api/admin/libraries/directories?' + new URLSearchParams({ path, after });
}

export function DirectoryPicker({ initial, close, select }: {
  initial: string;
  close: () => void;
  select: (path: string) => void;
}) {
  const [path, setPath] = useState(initial.startsWith('/') ? initial.replace(/\/+$/, '') || '/' : '');
  const [extra, setExtra] = useState<Folders>();
  const resource = useResource<Folders>(directoryURL(path));
  const action = useAction();
  const page = resource.data;
  const more = extra?.path === page?.path ? extra : undefined;
  const after = more?.next_after ?? page?.next_after;

  function navigate(next: string) {
    setExtra(undefined);
    action.clearError();
    setPath(next);
  }

  return (
    <Modal title="选择媒体目录" close={close}>
      <div className="library-directory-toolbar">
        <Button disabled={!page?.parent || action.busy} onClick={() => navigate(page?.parent ?? '')}>
          上一级
        </Button>
        <code>{page?.path ?? path}</code>
      </div>
      <ErrorMessage error={resource.error || action.error} />
      {resource.error && path && <Button onClick={() => navigate('')}>回到默认目录</Button>}
      {resource.loading && <Loading />}
      {page && (
        <List dense className="library-directory-list">
          {[...page.folders, ...(more?.folders ?? [])].map((folder) => (
            <ListItemButton key={folder.path} disabled={action.busy} onClick={() => navigate(folder.path)}>
              <ListItemIcon><Icon name="folder" /></ListItemIcon>
              <ListItemText primary={folder.name} />
            </ListItemButton>
          ))}
          {after && (
            <ListItemButton disabled={action.busy} onClick={() => action.run(async () => {
              const next = await api<Folders>(directoryURL(page.path, after));
              setExtra({ ...next, folders: [...(more?.folders ?? []), ...next.folders] });
            })}>加载更多</ListItemButton>
          )}
          {!page.folders.length && <li className="muted">没有子目录</li>}
        </List>
      )}
      <div className="modal-actions">
        <Button onClick={close}>取消</Button>
        <Button variant="contained" disabled={!page || resource.loading} onClick={() => page && select(page.path)}>
          选择此目录
        </Button>
      </div>
    </Modal>
  );
}
