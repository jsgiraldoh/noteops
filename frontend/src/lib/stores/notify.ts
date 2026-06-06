import { writable } from 'svelte/store';

export type NotifyType = 'success' | 'error' | 'info';

export interface Notification {
  id: number;
  type: NotifyType;
  message: string;
}

function createNotifyStore() {
  const { subscribe, update } = writable<Notification[]>([]);
  let seq = 0;

  function push(message: string, type: NotifyType, duration = 4000) {
    const id = seq++;
    update(n => [...n, { id, type, message }]);
    setTimeout(() => update(n => n.filter(x => x.id !== id)), duration);
  }

  return {
    subscribe,
    success: (msg: string) => push(msg, 'success'),
    error:   (msg: string) => push(msg, 'error', 5000),
    info:    (msg: string) => push(msg, 'info'),
    dismiss: (id: number)  => update(n => n.filter(x => x.id !== id))
  };
}

export const notify = createNotifyStore();
