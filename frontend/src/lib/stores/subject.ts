import { writable } from 'svelte/store';
import type { Subject } from '$lib/api/subjects';

export const currentSubject = writable<Subject | null>(null);
export const subjects = writable<Subject[]>([]);
