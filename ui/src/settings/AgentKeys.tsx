import { useState, type FormEvent } from 'react';
import { Select } from '../components/Select';
import { ErrorMessage, Loading } from '../components/common';
import { InfoTip } from '../components/InfoTip';
import { useToast } from '../components/Toast';
import { api } from '../lib/api';
import { useAction, useResource } from '../lib/hooks';
import { useLibraries } from '../libraries/LibraryProvider';

interface AgentKey {
  id: number;
  name: string;
  library_id: number;
  prefix: string;
  access: string;
  expires_at: number;
}

export function AgentKeys() {
  const { libraries } = useLibraries();
  const resource = useResource<{ keys: AgentKey[] }>('/api/admin/agent-keys');
  const action = useAction();
  const notify = useToast();
  const [library, setLibrary] = useState('');
  const [access, setAccess] = useState('read_write');
  const [days, setDays] = useState('90');
  const [created, setCreated] = useState<{ id: number; token: string } | null>(null);
  function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const name = String(new FormData(form).get('name')).trim();
    action.run(async () => {
      setCreated(null);
      const result = await api<{ key: AgentKey; token: string }>('/api/admin/agent-keys', {
        method: 'POST',
        body: { name, library_id: Number(library), access, expires_in_days: Number(days) },
      });
      setCreated({ id: result.key.id, token: result.token });
      form.reset();
      resource.reload();
    });
  }
  return (
    <section className="account-card agent-keys" aria-label="Agent API Key">
      <div className="device-card-heading">
        <h2>Agent API Key</h2>
        <InfoTip label="Agent API 用途">
          读取待识别媒体、提交 JSON 或 NFO 资料及上传海报；不开放播放、账户管理或文件删除权限。
        </InfoTip>
      </div>
      <a className="text-button" href="/api/admin/agent-keys/docs" target="_blank" rel="noopener noreferrer">
        接口文档
      </a>
      <form onSubmit={create}>
        <label>
          名称
          <input name="name" maxLength={80} required placeholder="例如：媒体识别 Agent" />
        </label>
        <label>
          媒体库
          <Select value={library} onValueChange={setLibrary}>
            <option value="">全部媒体库</option>
            {libraries.map((item) => (
              <option key={item.id} value={item.id}>
                {item.name}
              </option>
            ))}
          </Select>
        </label>
        <div className="agent-key-options">
          <label>
            权限
            <Select value={access} onValueChange={setAccess}>
              <option value="read_write">读取和写入资料</option>
              <option value="read">仅读取资料</option>
            </Select>
          </label>
          <label>
            有效期
            <Select value={days} onValueChange={setDays}>
              <option value="30">30 天</option>
              <option value="90">90 天</option>
              <option value="365">365 天</option>
            </Select>
          </label>
        </div>
        <button className="button primary" disabled={action.busy}>
          创建 Key
        </button>
      </form>
      {created && (
        <div className="agent-key-created">
          <label>
            API Key（仅显示一次）
            <input
              value={created.token}
              readOnly
              autoComplete="off"
              spellCheck={false}
              onFocus={(event) => event.currentTarget.select()}
            />
          </label>
          <div className="metadata-actions">
            <button
              className="text-button"
              onClick={() =>
                action.run(async () => {
                  try {
                    await navigator.clipboard.writeText(created.token);
                    notify('已复制');
                  } catch {
                    notify('点击 Key 输入框可选中复制');
                  }
                })
              }
            >
              复制
            </button>
            <button className="text-button" onClick={() => setCreated(null)}>
              收起
            </button>
          </div>
        </div>
      )}
      <ErrorMessage error={resource.error || action.error} />
      {resource.loading && !resource.data && <Loading />}
      <ul className="device-list">
        {resource.data?.keys.map((key) => (
          <li key={key.id}>
            <div>
              <strong>{key.name}</strong>
              <span>
                {key.library_id
                  ? libraries.find((item) => item.id === key.library_id)?.name || '指定媒体库'
                  : '全部媒体库'}{' '}
                · {key.access === 'read' ? '只读' : '读写'}
              </span>
              <span>
                {key.prefix} · {new Date(key.expires_at * 1000).toLocaleDateString()} 到期
              </span>
            </div>
            <button
              className="text-button"
              disabled={action.busy}
              onClick={() =>
                action.run(async () => {
                  await api('/api/admin/agent-keys/' + key.id, { method: 'DELETE' });
                  if (created?.id === key.id) setCreated(null);
                  resource.reload();
                })
              }
            >
              撤销
            </button>
          </li>
        ))}
      </ul>
    </section>
  );
}
