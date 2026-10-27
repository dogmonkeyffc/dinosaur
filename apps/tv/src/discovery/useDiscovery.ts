import { useEffect, useState } from 'react';
import Zeroconf, { type Service } from 'react-native-zeroconf';
import type { ConnectedServer } from '@dinosaur/contracts';
import { probeServer } from '@dinosaur/native-kit';

export function useDiscovery() {
  const [servers, setServers] = useState<ConnectedServer[]>([]);
  const [error, setError] = useState('');
  const [scanning, setScanning] = useState(true);
  const [revision, setRevision] = useState(0);
  useEffect(() => {
    let active = true;
    const resolving = new Set<string>();
    const discovery = new Zeroconf();
    setScanning(true);
    setError('');
    async function resolved(service: Service) {
      const key = service.fullName || service.name;
      if (resolving.has(key)) return;
      resolving.add(key);
      const candidates = [
        service.txt?.url,
        ...(service.ipv4 || service.addresses || [])
          .filter((ip) => /^\d+\.\d+\.\d+\.\d+$/.test(ip))
          .map((ip) => `http://${ip}:${service.port}`),
      ].filter((url): url is string => Boolean(url));
      for (const url of candidates) {
        try {
          const server = await probeServer(url);
          if (!active) return;
          if (service.txt?.id && service.txt.id !== server.id) continue;
          setServers((items) => [...items.filter((item) => item.id !== server.id), server]);
          return;
        } catch {
          /* Try the advertised LAN address if the public URL cannot be reached. */
        }
      }
      resolving.delete(key);
    }
    discovery.on('resolved', resolved);
    discovery.on('error', () => {
      if (active) setError('无法扫描局域网，可重新扫描或手动连接。');
    });
    try {
      discovery.scan({ type: 'dinosaur', protocol: 'tcp', domain: 'local.', implType: 'NSD' });
    } catch {
      setError('局域网发现不可用，请手动连接。');
    }
    const timer = setTimeout(() => {
      if (active) setScanning(false);
    }, 8000);
    return () => {
      active = false;
      clearTimeout(timer);
      discovery.stop();
      discovery.removeAllListeners();
    };
  }, [revision]);
  return {
    servers,
    error,
    scanning,
    restart: () => {
      setServers([]);
      setRevision((value) => value + 1);
    },
  };
}
