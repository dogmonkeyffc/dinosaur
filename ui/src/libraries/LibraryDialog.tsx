import { FormInput } from '../components/FormInput';
import { Button, FormControlLabel, Radio, RadioGroup } from '@mui/material';
import { useState, type FormEvent } from 'react';
import { api } from '../lib/api';
import { useAction } from '../lib/hooks';
import type { Library, MediaKind } from '../lib/types';
import { ErrorMessage, Modal } from '../components/common';
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
          <FormInput
            className="form-control"
            name="name"
            required
            maxLength={80}
            defaultValue={item?.name || ''}
          />
        </label>
        <fieldset className="type-fieldset">
          <legend>内容类型</legend>
          <RadioGroup
            name="type"
            value={type}
            row
            onChange={(_, value) => setType(value === 'tv' ? 'tv' : 'movie')}
          >
            <FormControlLabel value="movie" control={<Radio />} label="电影" />
            <FormControlLabel value="tv" control={<Radio />} label="电视剧" />
          </RadioGroup>
        </fieldset>
        <label htmlFor="library-path">
          媒体目录
          <FormInput
            className="form-control path-input"
            id="library-path"
            name="path"
            required
            maxLength={4096}
            defaultValue={item?.root_path || ''}
            placeholder="/mnt/media/movies"
          />
        </label>
        <ErrorMessage error={action.error} />
        <div className="modal-actions">
          <Button variant="outlined" type="button" onClick={close}>
            取消
          </Button>
          <Button variant="contained" disabled={action.busy} type="submit">
            保存媒体库
          </Button>
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
        <Button variant="outlined" onClick={close}>
          取消
        </Button>
        <Button
          variant="contained"
          color="error"
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
        </Button>
      </div>
    </Modal>
  );
}
