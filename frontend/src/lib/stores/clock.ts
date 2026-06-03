import { writable } from 'svelte/store';

export interface SessionTick {
  session_id: string;
  elapsed_sec: number;
  remaining_sec: number;
  duration_min: number;
  is_active: boolean;
}

export const clockStore = writable<SessionTick | null>(null);

let ws: WebSocket | null = null;

export function connectClock(sessionId: string) {
  const WS_URL = import.meta.env.PUBLIC_WS_URL ?? 'ws://localhost:8080';
  disconnectClock();
  ws = new WebSocket(`${WS_URL}/ws/session/${sessionId}`);
  ws.onmessage = (e) => {
    try { clockStore.set(JSON.parse(e.data)); } catch {}
  };
  ws.onclose = () => clockStore.set(null);
}

export function disconnectClock() {
  if (ws) { ws.close(); ws = null; }
  clockStore.set(null);
}
