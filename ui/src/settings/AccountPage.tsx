import type { FormEvent } from 'react';
import { useAuth } from '../auth/AuthProvider';
import { ErrorMessage, Heading } from '../components/common';
import { useToast } from '../components/Toast';
import { api } from '../lib/api';
import { useAction } from '../lib/hooks';

export function AccountPage() {
  const auth = useAuth();
  const action = useAction();
  const notify = useToast();
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const values = new FormData(event.currentTarget);
    action.run(async () => {
      if (values.get('new_password') !== values.get('confirm_password'))
        throw new Error('两次输入的新密码不一致。');
      await api('/api/auth/password', {
        method: 'POST',
        body: { current_password: values.get('current_password'), new_password: values.get('new_password') },
      });
      auth.expired();
      notify('密码已更新，请重新登录。');
    });
  }
  return (
    <section id="account-section">
      <Heading title="账户安全" />
      <div className="account-card">
        <h2>修改密码</h2>
        <form id="password-form" onSubmit={submit}>
          <label>
            当前密码
            <input
              name="current_password"
              type="password"
              autoComplete="current-password"
              required
              maxLength={1024}
            />
          </label>
          <label>
            新密码
            <input
              name="new_password"
              type="password"
              autoComplete="new-password"
              required
              maxLength={1024}
            />
            <span className="field-help">至少 12 字节，建议使用较长的密码短语。</span>
          </label>
          <label>
            确认新密码
            <input
              name="confirm_password"
              type="password"
              autoComplete="new-password"
              required
              maxLength={1024}
            />
          </label>
          <ErrorMessage error={action.error} />
          <button className="button primary" disabled={action.busy}>
            更新密码
          </button>
        </form>
      </div>
    </section>
  );
}
