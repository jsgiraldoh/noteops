import { api } from './client';

export interface Student { id: string; full_name: string; email: string; code: string; created_at: string; }

export const studentsApi = {
  create: (data: { full_name: string; email: string; code?: string }) =>
    api.post<Student>('/students', data),
  bySubject: (subjectId: string) =>
    api.get<Student[]>(`/subjects/${subjectId}/students`),
  enroll: (subjectId: string, studentId: string) =>
    api.post(`/subjects/${subjectId}/enroll`, { student_id: studentId }),
  update: (id: string, data: { full_name: string; email: string; code?: string }) =>
    api.patch<Student>(`/students/${id}`, data)
};
