import { Button, Card } from '@mui/material';
import { useState, type FormEvent } from 'react';
import { ErrorMessage } from '../components/common';
import { FormInput } from '../components/FormInput';
import { api } from '../lib/api';
import { useAction } from '../lib/hooks';

interface TVPairing {
  pairing_id: string;
  name: string;
}

export function TVLogin() {
  const [code, setCode] = useState('');
  const [pairing, setPairing] = useState<TVPairing>();
  const [approved, setApproved] = useState(false);
  const action = useAction();
  const validCode = /^[0-9]{6}$/.test(code);

  function lookup(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!validCode || pairing) return;
    void action.run(async () => {
      setApproved(false);
      setPairing(await api<TVPairing>('/api/devices/pairings/lookup', {
        method: 'POST',
        body: { code },
      }));
    });
  }

  function approve() {
    if (!pairing) return;
    void action.run(async () => {
      await api('/api/devices/pairings/approve', {
        method: 'POST',
        body: { pairing_id: pairing.pairing_id },
      });
      setPairing(undefined);
      setCode('');
      setApproved(true);
    });
  }

  return (
    <Card component="section" className="account-card device-card">
      <h2>登录电视</h2>
      {pairing ? (
        <>
          <p>允许「{pairing.name}」登录当前账号？</p>
          <div className="metadata-actions">
            <Button variant="contained" disabled={action.busy} onClick={approve}>
              确认登录
            </Button>
            <Button
              disabled={action.busy}
              onClick={() => {
                setPairing(undefined);
                action.clearError();
              }}
            >
              取消
            </Button>
          </div>
        </>
      ) : (
        <form className="metadata-actions" onSubmit={lookup}>
          <label>
            电视上的 6 位配对码
            <FormInput
              value={code}
              onChange={(event) => {
                setCode(event.target.value.replace(/[^0-9]/g, ''));
                setApproved(false);
                action.clearError();
              }}
              inputMode="numeric"
              autoComplete="off"
              maxLength={6}
              disabled={action.busy}
            />
          </label>
          <Button variant="contained" type="submit" disabled={!validCode || action.busy}>
            查找电视
          </Button>
        </form>
      )}
      {approved && <p role="status">已授权，请返回电视继续。</p>}
      <ErrorMessage error={action.error} />
    </Card>
  );
}
