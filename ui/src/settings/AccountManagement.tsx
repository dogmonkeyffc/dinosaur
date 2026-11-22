import { Button, Card } from '@mui/material';
import { useState } from 'react';
import { ErrorMessage, Loading, Modal } from '../components/common';
import { FormInput } from '../components/FormInput';
import { api } from '../lib/api';
import { useAction, useResource } from '../lib/hooks';
import type { User } from '../lib/types';

type Account = User & { disabled: boolean };
const endpoint = '/api/admin/users';

export function AccountManagement({ currentID }: { currentID: number }) {
  const resource = useResource<{ users: Account[] }>(endpoint);
  const action = useAction();
  const [creating, setCreating] = useState(false);
  const [changing, setChanging] = useState<Account>();
  const close = () => {
    if (action.busy) return;
    setCreating(false);
    setChanging(undefined);
    action.clearError();
  };

  return (
    <Card component="section" className="account-card">
      <div className="device-card-heading">
        <h2>账号管理</h2>
        <Button variant="outlined" onClick={() => setCreating(true)}>添加账号</Button>
      </div>
      <ErrorMessage error={resource.error} />
      {resource.error && <Button onClick={resource.reload}>重试</Button>}
      {resource.loading && !resource.data && <Loading />}
      <ul className="device-list">
        {resource.data?.users.map((user) => (
          <li key={user.id}>
            <div>
              <strong>{user.username}{user.id === currentID ? '（当前账号）' : ''}</strong>
              <span>{user.role === 'admin' ? '管理员' : '普通账号'}{user.disabled ? ' · 已停用' : ''}</span>
            </div>
            {user.id !== currentID && (
              <Button onClick={() => setChanging(user)}>{user.disabled ? '启用' : '停用'}</Button>
            )}
          </li>
        ))}
      </ul>
      {creating && (
        <Modal title="添加账号" close={close}>
          <form onSubmit={(event) => {
            event.preventDefault();
            const values = new FormData(event.currentTarget);
            const username = String(values.get('username')).trim().toLowerCase();
            const password = String(values.get('password'));
            void action.run(async () => {
              if (!/^[a-z0-9][a-z0-9_.-]{2,31}$/.test(username))
                throw new Error('用户名须为 3 至 32 个字符，以字母或数字开头，可包含点、连字符和下划线。');
              if (new TextEncoder().encode(password).length < 12)
                throw new Error('密码至少需要 12 字节。');
              await api(endpoint, {
                method: 'POST',
                body: { username, password },
              });
              setCreating(false);
              resource.reload();
            });
          }}>
            <label>用户名<FormInput name="username" autoComplete="off" maxLength={32} required /></label>
            <label>密码（至少 12 字节）<FormInput name="password" type="password" autoComplete="new-password" required /></label>
            <ErrorMessage error={action.error} />
            <div className="modal-actions">
              <Button onClick={close} disabled={action.busy}>取消</Button>
              <Button type="submit" variant="contained" disabled={action.busy}>创建账号</Button>
            </div>
          </form>
        </Modal>
      )}
      {changing && (
        <Modal title={`${changing.disabled ? '启用' : '停用'}「${changing.username}」？`} close={close}>
          <ErrorMessage error={action.error} />
          <div className="modal-actions">
            <Button onClick={close} disabled={action.busy}>取消</Button>
            <Button variant="contained" disabled={action.busy} onClick={() => action.run(async () => {
              await api(`${endpoint}/${changing.id}`, { method: 'PATCH', body: { disabled: !changing.disabled } });
              setChanging(undefined);
              resource.reload();
            })}>{changing.disabled ? '确认启用' : '确认停用'}</Button>
          </div>
        </Modal>
      )}
    </Card>
  );
}
