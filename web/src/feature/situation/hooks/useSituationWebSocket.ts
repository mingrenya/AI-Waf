import { useEffect, useRef, useCallback, useState } from 'react';
import type { WSSituationMessage } from '@/types/situation';

const WS_BASE = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}`;

export function useSituationWebSocket() {
  const wsRef = useRef<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const handlersRef = useRef<Map<string, Set<(payload: unknown) => void>>>(new Map());

  const subscribe = useCallback((type: string, handler: (payload: unknown) => void) => {
    if (!handlersRef.current.has(type)) handlersRef.current.set(type, new Set());
    handlersRef.current.get(type)!.add(handler);
    return () => { handlersRef.current.get(type)?.delete(handler); };
  }, []);

  useEffect(() => {
    const token = localStorage.getItem('token');
    const ws = new WebSocket(`${WS_BASE}/api/v1/ws?token=${token}`);
    wsRef.current = ws;

    ws.onopen = () => setConnected(true);
    ws.onclose = () => setConnected(false);
    ws.onmessage = (e) => {
      try {
        const msg: WSSituationMessage = JSON.parse(e.data);
        handlersRef.current.get(msg.type)?.forEach(h => h(msg.payload));
      } catch { /* ignore parse errors */ }
    };

    return () => ws.close();
  }, []);

  return { connected, subscribe };
}
