import { api } from './client';

export interface Session { id: string; subject_id: string; starts_at: string; duration_min: number; slot_min: number; room: string; active: boolean; }
export interface Slot { id: string; session_id: string; number: number; starts_at: string; duration_min: number; student_id: string | null; reserved_at: string | null; }

export const sessionsApi = {
  create: (data: { subject_id: string; starts_at: string; duration_min: number; slot_min: number; room?: string }) =>
    api.post<{ session: Session; slots: Slot[] }>('/sessions', data),
  activate: (sessionId: string) =>
    api.post(`/sessions/${sessionId}/activate`, {}),
  slots: (sessionId: string) =>
    api.get<Slot[]>(`/sessions/${sessionId}/slots`),
  reserve: (sessionId: string, slotId: string, studentId: string) =>
    api.post<Slot>(`/sessions/${sessionId}/slots/${slotId}/reserve`, { student_id: studentId })
};
