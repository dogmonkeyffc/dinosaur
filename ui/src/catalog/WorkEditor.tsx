import { Button, Dialog, DialogActions, DialogContent, DialogTitle } from '@mui/material';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { FormInput } from '../components/FormInput';
import { ErrorMessage, Loading, Pagination } from '../components/common';
import { api } from '../lib/api';
import { useAction, useDebounced, useResource } from '../lib/hooks';
import type { Page, Series, Work, WorkDetail } from '../lib/types';
import { WorkMembers } from './WorkMembers';

export function WorkEditor({
  libraryID,
  detail,
  close,
  saved,
}: {
  libraryID: number;
  detail?: WorkDetail;
  close: () => void;
  saved: (work: Work) => void;
}) {
  const navigate = useNavigate();
  const [title, setTitle] = useState(detail?.work.title || '');
  const [members, setMembers] = useState<Series[]>(detail?.all_members || detail?.members || []);
  const [query, setQuery] = useState('');
  const [page, setPage] = useState(1);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const search = useDebounced(query);
  const action = useAction();
  const params = new URLSearchParams({
    raw: '1',
    library_id: String(libraryID),
    page: String(page),
    page_size: '24',
    q: search,
  });
  const candidates = useResource<Page<Series>>(`/api/series?${params}`);
  const work = detail?.work;
  function save() {
    action.run(async () => {
      const result = await api<{ work: Work }>(`/api/admin/works${work ? `/${work.id}` : ''}`, {
        method: work ? 'PUT' : 'POST',
        body: { title: title.trim(), series_ids: members.map((item) => item.id) },
      });
      saved(result.work);
    });
  }
  return (
    <Dialog
      open
      onClose={action.busy ? undefined : close}
      fullWidth
      maxWidth="md"
      aria-labelledby="work-editor-title"
    >
      <DialogTitle id="work-editor-title">{work ? '编辑作品分组' : '新建作品分组'}</DialogTitle>
      <DialogContent>
        <fieldset disabled={action.busy} className="work-editor-fields">
          <label>
            作品名称
            <FormInput
              value={title}
              onChange={(event) => setTitle(event.target.value)}
              maxLength={128}
              required
            />
          </label>
          <WorkMembers members={members} setMembers={setMembers} />
          <label>
            搜索篇章
            <FormInput
              type="search"
              value={query}
              onChange={(event) => {
                setQuery(event.target.value);
                setPage(1);
              }}
              maxLength={128}
            />
          </label>
          <ErrorMessage error={candidates.error} />
          {candidates.loading && <Loading />}
          <ul className="work-candidates">
            {candidates.data?.items.map((item) => {
              const selected = members.some((member) => member.id === item.id);
              const other = !!item.work_id && item.work_id !== work?.id;
              return (
                <li key={item.id}>
                  <span>{item.title}</span>
                  <Button disabled={selected || other} onClick={() => setMembers([...members, item])}>
                    {selected ? '已添加' : other ? '已归入其他作品' : '添加'}
                  </Button>
                </li>
              );
            })}
          </ul>
          {!candidates.loading && candidates.data?.items.length === 0 && <p>没有找到篇章</p>}
          <Pagination page={page} total={candidates.data?.total || 0} size={24} onChange={setPage} />
        </fieldset>
        <ErrorMessage error={action.error} />
        {work && (
          <div className="movie-play-actions">
            <Button color="error" disabled={action.busy} onClick={() => setConfirmDelete(true)}>
              解散分组
            </Button>
            {confirmDelete && (
              <>
                <span>解散后各篇章恢复独立显示，文件与观看记录保留。</span>
                <Button
                  color="error"
                  disabled={action.busy}
                  onClick={() =>
                    action.run(async () => {
                      await api(`/api/admin/works/${work.id}`, { method: 'DELETE' });
                      close();
                      navigate(`/series?library_id=${libraryID}`, { replace: true });
                    })
                  }
                >
                  确认解散
                </Button>
                <Button disabled={action.busy} onClick={() => setConfirmDelete(false)}>
                  取消解散
                </Button>
              </>
            )}
          </div>
        )}
      </DialogContent>
      <DialogActions>
        <Button disabled={action.busy} onClick={close}>
          取消
        </Button>
        <Button variant="contained" disabled={action.busy || !title.trim() || !members.length} onClick={save}>
          保存
        </Button>
      </DialogActions>
    </Dialog>
  );
}
