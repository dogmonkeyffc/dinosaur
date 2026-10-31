import { useEffect, useState } from 'react';
import { QRCodeSVG } from 'qrcode.react';
import { useAuth } from '../auth/AuthProvider';
import { api } from '../lib/api';
import { useAction } from '../lib/hooks';
import { ErrorMessage } from '../components/common';

interface Ticket {
  qr_payload: string;
  expires_at: number;
}

export function MobileLogin() {
  const { user } = useAuth();
  const [ticket, setTicket] = useState<Ticket>();
  const [now, setNow] = useState(Date.now());
  const action = useAction();
  useEffect(() => {
    if (!ticket) return;
    const timer = window.setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(timer);
  }, [ticket]);
  const seconds = ticket ? Math.max(0, Math.ceil(ticket.expires_at - now / 1000)) : 0;
  return (
    <section className="account-card device-card">
      <div className="device-card-heading">
        <h2>手机扫码登录</h2>
        <button
          className="button secondary"
          disabled={action.busy}
          onClick={() =>
            action.run(async () => {
              const result = await api<Ticket>('/api/devices/qr', { method: 'POST', body: {} });
              setNow(Date.now());
              setTicket(result);
            })
          }
        >
          {ticket ? '刷新二维码' : '生成二维码'}
        </button>
      </div>
      <span>登录账号：{user?.username}</span>
      {ticket && seconds > 0 && (
        <div className="mobile-login-qr">
          <QRCodeSVG
            value={ticket.qr_payload}
            size={224}
            marginSize={4}
            level="M"
            title="使用 Dinosaur 手机 App 扫码登录"
          />
          <span role="status">{seconds} 秒后过期</span>
        </div>
      )}
      {ticket && !seconds && <p role="status">二维码已过期</p>}
      <ErrorMessage error={action.error} />
    </section>
  );
}
