import { writable } from 'svelte/store';
import { PUBLIC_WS_URL } from '$env/static/public';

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
  disconnectClock();
  const WS_URL = PUBLIC_WS_URL || 'ws://noteops.local';
  const socket = new WebSocket(`${WS_URL}/ws/session/${sessionId}`);
  ws = socket;
  socket.onmessage = (e) => {
    try { clockStore.set(JSON.parse(e.data)); } catch {}
  };
  // Solo resetea el store si este socket sigue siendo el activo.
  // Evita que el onclose de un socket reemplazado anule el clock del nuevo.
  socket.onclose = () => {
    if (ws === socket) clockStore.set(null);
  };
}

export function disconnectClock() {
  if (ws) {
    ws.onclose = null; // Previene que el cierre intencional dispare el reset del store
    ws.close();
    ws = null;
  }
  clockStore.set(null);
}
