import { api } from './client';

export interface Subject {
  id: string;
  name: string;
  period: string;
  group_name: string;
  faculty: string;
  teacher_id: string;
}

export interface CreateSubjectData {
  name: string;
  period: string;
  group_name?: string;
  faculty?: string;
}

export const subjectsApi = {
  list: () => api.get<Subject[]>('/subjects'),
  create: (data: CreateSubjectData) => api.post<Subject>('/subjects', data),
  remove: (id: string) => api.delete<{ deleted: boolean }>(`/subjects/${id}`)
};
