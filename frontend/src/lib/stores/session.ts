import { writable } from 'svelte/store';
import type { Session } from '$lib/api/sessions';

// Sesión en curso (creada pero no finalizada).
// Mientras exista, el cambio de materia queda bloqueado.
export const activeSession = writable<Session | null>(null);
