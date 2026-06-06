import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import type { Session } from '$lib/api/sessions';

const KEY = 'noteops_session';

function createSessionStore() {
  const initial: Session | null = browser
    ? (() => { try { const v = localStorage.getItem(KEY); return v ? JSON.parse(v) : null; } catch { return null; } })()
    : null;

  const store = writable<Session | null>(initial);

  return {
    subscribe: store.subscribe,
    set(v: Session | null) {
      if (browser) {
        if (v) localStorage.setItem(KEY, JSON.stringify(v));
        else localStorage.removeItem(KEY);
      }
      store.set(v);
    },
    update(fn: (v: Session | null) => Session | null) {
      store.update(current => {
        const next = fn(current);
        if (browser) {
          if (next) localStorage.setItem(KEY, JSON.stringify(next));
          else localStorage.removeItem(KEY);
        }
        return next;
      });
    }
  };
}

export const activeSession = createSessionStore();
