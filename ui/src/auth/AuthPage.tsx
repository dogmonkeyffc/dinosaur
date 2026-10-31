import { InfoTip } from '../components/InfoTip';
import { useState, type FormEvent } from 'react';
import { useAuth } from './AuthProvider';
import { api } from '../lib/api';
import { useAction } from '../lib/hooks';
import { Brand, ErrorMessage, Icon } from '../components/common';

export function AuthPage() {
  const auth = useAuth();
  const setup = auth.status === 'setup';
  const action = useAction();
  const [visible, setVisible] = useState(false);
  const [notice, setNotice] = useState('');
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const values = new FormData(form);
    const username = String(values.get('username')).trim();
    const password = String(values.get('password'));
    await action.run(async () => {
      if (setup && password !== values.get('confirm')) throw new Error('两次输入的密码不一致。');
      if (setup && new TextEncoder().encode(password).length < 12) throw new Error('密码至少需要 12 字节。');
      if (setup) {
        await api('/api/setup', {
          method: 'POST',
          body: { username, password },
          headers: { 'X-Setup-Token': String(values.get('token')).trim() },
        });
        form.reset();
        auth.initialized();
        setNotice('管理员已创建，请使用刚才的密码登录。');
      } else await auth.login(username, password);
    });
  }
  return (
    <main id="auth-screen" className="auth-screen">
      <section className="auth-panel">
        <div className="auth-form-wrap">
          <Brand />
          <h1>{setup ? '创建管理员' : '登录'}</h1>
          {notice && <p role="status">{notice}</p>}
          <form id="auth-form" onSubmit={submit}>
            {setup && (
              <label htmlFor="setup-token">
                <span className="field-label">
                  <span id="setup-token-label">初始化令牌</span>
                  <InfoTip label="初始化令牌位置">
                    默认位于 ~/.cache/dinosaur/setup-token；自定义路径见服务启动日志。
                  </InfoTip>
                </span>
                <input
                  aria-labelledby="setup-token-label"
                  id="setup-token"
                  name="token"
                  type="password"
                  autoComplete="off"
                  required
                />
              </label>
            )}
            <label>
              用户名
              <input
                id="username"
                name="username"
                autoComplete="username"
                required
                maxLength={32}
                pattern={'[A-Za-z0-9][A-Za-z0-9_.\\-]{2,31}'}
              />
            </label>
            <label>
              <span id="password-label">{setup ? '密码（至少 12 字节）' : '密码'}</span>
              <span className="password-field">
                <input
                  aria-labelledby="password-label"
                  id="password"
                  name="password"
                  type={visible ? 'text' : 'password'}
                  autoComplete={setup ? 'new-password' : 'current-password'}
                  required
                  maxLength={1024}
                />
                <button
                  type="button"
                  className="show-password"
                  onClick={() => setVisible(!visible)}
                  aria-label={visible ? '隐藏密码' : '显示密码'}
                >
                  {visible ? '隐藏' : '显示'}
                </button>
              </span>
            </label>
            {setup && (
              <label>
                确认密码
                <input name="confirm" type="password" autoComplete="new-password" required maxLength={1024} />
              </label>
            )}
            <ErrorMessage error={action.error} />
            <button className="button primary auth-submit" type="submit" disabled={action.busy}>
              <span>{action.busy ? '请稍候…' : setup ? '创建管理员' : '登录'}</span>
              <Icon name="arrow" />
            </button>
          </form>
        </div>
      </section>
    </main>
  );
}
