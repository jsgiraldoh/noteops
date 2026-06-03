import { api } from './client';
export interface Subject { id: string; name: string; period: string; group_name: string; faculty: string; teacher_id: string; }
export const subjectsApi = {
  list: () => api.get<Subject[]>('/subjects')
};
