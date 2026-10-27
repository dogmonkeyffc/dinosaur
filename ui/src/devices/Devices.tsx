import { ErrorMessage, Loading } from '../components/common';
import { useAction, useResource } from '../lib/hooks';
import { api } from '../lib/api';

interface Device {
  id: string;
  name: string;
  current: boolean;
  expires_at: number;
}

export function Devices() {
  const resource = useResource<{ devices: Device[] }>('/api/devices', 10000);
  const action = useAction();
  return (
    <section className="account-card device-card">
      <div className="device-card-heading">
        <h2>已登录设备</h2>
        <button className="text-button" onClick={resource.reload}>
          刷新
        </button>
      </div>
      {resource.loading && !resource.data && <Loading />}
      <ErrorMessage error={resource.error || action.error} />
      <ul className="device-list">
        {resource.data?.devices.map((device) => (
          <li key={device.id}>
            <div>
              <strong>
                {device.name}
                {device.current ? '（当前设备）' : ''}
              </strong>
              <span>{new Date(device.expires_at * 1000).toLocaleString()} 到期</span>
            </div>
            <button
              className="text-button"
              disabled={action.busy}
              onClick={() =>
                action.run(async () => {
                  await api('/api/devices/' + device.id, { method: 'DELETE' });
                  resource.reload();
                })
              }
            >
              退出设备
            </button>
          </li>
        ))}
      </ul>
      {resource.data && !resource.data.devices.length && <p className="muted">暂无 App 登录设备</p>}
    </section>
  );
}
