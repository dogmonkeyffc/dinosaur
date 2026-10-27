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
      <section className="auth-story">
        <Brand />
        <div className="story-content">
          <span className="eyebrow">YOUR COLLECTION. YOUR SPACE.</span>
          <h1>
            让好故事，
            <br />
            有自己的家<span>。</span>
          </h1>
          <div className="collection-art" aria-hidden="true">
            <div className="art-grid" />
            <div className="art-card art-back">
              <Icon name="tv" />
            </div>
            <div className="art-card art-front">
              <Icon name="film" />
              <div className="art-lines">
                <span />
                <span />
              </div>
              <span className="art-caption">THE PERSONAL COLLECTION</span>
            </div>
            <span className="art-orbit" />
            <span className="art-spark">✳</span>
          </div>
        </div>
      </section>
      <section className="auth-panel">
        <div className="auth-form-wrap">
          <span className="small-label">{setup ? 'MAKE YOURSELF AT HOME' : 'WELCOME BACK'}</span>
          <h2>{setup ? '开启你的私人空间' : '欢迎回来'}</h2>
          {notice && <p role="status">{notice}</p>}
          <form id="auth-form" onSubmit={submit}>
            {setup && (
              <label>
                初始化令牌
                <input
                  name="token"
                  type="password"
                  autoComplete="off"
                  required
                  placeholder="粘贴服务器的初始化令牌"
                />
                <span className="field-help">
                  默认位于 ~/.cache/dinosaur/setup-token；自定义路径见服务启动日志。
                </span>
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
                placeholder="输入你的用户名"
              />
            </label>
            <label>
              密码
              <span className="password-field">
                <input
                  id="password"
                  name="password"
                  type={visible ? 'text' : 'password'}
                  autoComplete={setup ? 'new-password' : 'current-password'}
                  required
                  maxLength={1024}
                  placeholder="输入密码"
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
                <input
                  name="confirm"
                  type="password"
                  autoComplete="new-password"
                  required
                  maxLength={1024}
                  placeholder="再次输入密码"
                />
              </label>
            )}
            <ErrorMessage error={action.error} />
            <button className="button primary auth-submit" type="submit" disabled={action.busy}>
              <span>{action.busy ? '请稍候…' : setup ? '创建管理员' : '登录'}</span>
              <Icon name="arrow" />
            </button>
          </form>
        </div>
        <span className="panel-footer">Dinosaur Media Server</span>
      </section>
    </main>
  );
}
