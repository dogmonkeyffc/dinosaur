import { useEffect, useState } from 'react';
import Zeroconf, { type Service } from 'react-native-zeroconf';
import type { ConnectedServer } from '@dinosaur/contracts';
import { probeServer } from './probeServer';

export function useDiscovery() {
  const [servers, setServers] = useState<ConnectedServer[]>([]);
  const [error, setError] = useState('');
  const [scanning, setScanning] = useState(true);
  const [revision, setRevision] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    const resolving = new Set<string>();
    let discovery: Zeroconf | undefined;
    setScanning(true);
    setError('');
    async function resolved(service: Service) {
      const key = service.fullName || service.name;
      if (resolving.has(key) || controller.signal.aborted) return;
      resolving.add(key);
      const candidates = [
        service.txt?.url,
        ...(service.ipv4 || service.addresses || [])
          .filter((ip) => /^\d+\.\d+\.\d+\.\d+$/.test(ip))
          .map((ip) => `http://${ip}:${service.port}`),
      ].filter((url): url is string => Boolean(url));
      try {
        for (const url of candidates) {
          try {
            const server = await probeServer(url, controller.signal);
            if (controller.signal.aborted) return;
            if (service.txt?.id && service.txt.id !== server.id) continue;
            setServers((items) => [...items.filter((item) => item.id !== server.id), server]);
            return;
          } catch {
            if (controller.signal.aborted) return;
          }
        }
      } finally {
        resolving.delete(key);
      }
    }
    try {
      discovery = new Zeroconf();
      discovery.on('resolved', resolved);
      discovery.on('error', () => {
        if (!controller.signal.aborted) setError('无法扫描局域网，请手动连接或重新扫描。');
      });
      discovery.scan({ type: 'dinosaur', protocol: 'tcp', domain: 'local.', implType: 'NSD' });
    } catch {
      setError('局域网发现不可用，请手动连接。');
    }
    const timer = setTimeout(() => {
      if (!controller.signal.aborted) setScanning(false);
    }, 8000);
    return () => {
      controller.abort();
      clearTimeout(timer);
      try {
        discovery?.stop();
      } catch {
        /* Native discovery may already be unavailable; cancellation still completes. */
      } finally {
        discovery?.removeAllListeners();
        discovery?.removeDeviceListeners();
      }
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
