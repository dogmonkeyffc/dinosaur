import { InfoTip } from '../components/InfoTip';
import { useState, type FormEvent } from 'react';
import { api } from '../lib/api';
import { useAction } from '../lib/hooks';
import type { Library, MediaKind } from '../lib/types';
import { ErrorMessage, Icon, Modal } from '../components/common';
import { useLibraries } from './LibraryProvider';

export function LibraryDialog({ item, close }: { item?: Library; close: () => void }) {
  const [type, setType] = useState<MediaKind>(item?.type || 'movie');
  const action = useAction();
  const { reload } = useLibraries();
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const values = new FormData(event.currentTarget);
    action.run(async () => {
      const body = { name: String(values.get('name')).trim(), root_path: values.get('path'), type };
      await api(item ? `/api/admin/libraries/${item.id}` : '/api/admin/libraries', {
        method: item ? 'PUT' : 'POST',
        body,
      });
      reload();
      close();
    });
  }
  return (
    <Modal title={item ? '编辑媒体库' : '添加媒体库'} close={close}>
      <form id="library-form" onSubmit={submit}>
        <label>
          媒体库名称
          <input name="name" required maxLength={80} defaultValue={item?.name || ''} />
        </label>
        <fieldset className="type-fieldset">
          <legend>内容类型</legend>
          <div className="type-options">
            {(['movie', 'tv'] as const).map((value) => (
              <label key={value} className="type-option">
                <input
                  type="radio"
                  name="type"
                  value={value}
                  checked={type === value}
                  onChange={() => setType(value)}
                />
                <Icon name={value === 'tv' ? 'tv' : 'film'} />
                <span>
                  <strong>{value === 'tv' ? '电视剧' : '电影'}</strong>
                </span>
              </label>
            ))}
          </div>
        </fieldset>
        <label htmlFor="library-path">
          <span className="field-label">
            <span id="library-path-label">媒体目录</span>
            <InfoTip label="媒体目录要求">填写服务器上的绝对路径，目录需已存在且可读取。</InfoTip>
          </span>
          <input
            aria-labelledby="library-path-label"
            id="library-path"
            name="path"
            className="path-input"
            required
            maxLength={4096}
            defaultValue={item?.root_path || ''}
            placeholder="/mnt/media/movies"
          />
        </label>
        <ErrorMessage error={action.error} />
        <div className="modal-actions">
          <button className="button secondary" type="button" onClick={close}>
            取消
          </button>
          <button className="button primary" disabled={action.busy}>
            保存媒体库
          </button>
        </div>
      </form>
    </Modal>
  );
}

export function DeleteLibrary({ item, close }: { item: Library; close: () => void }) {
  const action = useAction();
  const { reload } = useLibraries();
  return (
    <Modal title="移除媒体库？" close={close}>
      <p className="muted">将移除「{item.name}」的配置，原目录和所有媒体文件都会保留。</p>
      <ErrorMessage error={action.error} />
      <div className="modal-actions">
        <button className="button secondary" onClick={close}>
          取消
        </button>
        <button
          className="button danger"
          disabled={action.busy}
          onClick={() =>
            action.run(async () => {
              await api(`/api/admin/libraries/${item.id}`, { method: 'DELETE' });
              reload();
              close();
            })
          }
        >
          确认移除
        </button>
      </div>
    </Modal>
  );
}
