import { writable, derived } from 'svelte/store';
import type { User } from '$lib/api/auth';

export const user = writable<User | null>(null);
export const isLoggedIn = derived(user, $u => $u !== null);
