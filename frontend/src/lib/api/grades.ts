import { api } from './client';

export interface Grade { id: string; enrollment_id: string; activity_id: string; value: number | null; comment: string; recorded_at: string; }
export interface Activity { id: string; cut_id: string; name: string; weight: number; }
export interface Cut { id: string; subject_id: string; number: number; name: string; weight: number; activities: Activity[]; }
export interface FinalGrade { enrollment_id: string; student_id: string; subject_id: string; final_grade: number; }
export interface Enrollment { id: string; student_id: string; subject_id: string; }
export interface SubjectGrades {
  cuts: Cut[];
  students: { id: string; full_name: string; email: string }[];
  enrollments: Enrollment[];
  grades: Grade[];
  final_grades: FinalGrade[];
}

export const gradesApi = {
  record: (data: { enrollment_id: string; activity_id: string; value: number; comment?: string }) =>
    api.post<Grade>('/grades', data),
  updateComment: (gradeId: string, comment: string) =>
    api.patch(`/grades/${gradeId}/comment`, { comment }),
  bySubject: (subjectId: string) =>
    api.get<SubjectGrades>(`/subjects/${subjectId}/grades`),
  finalBySubject: (subjectId: string) =>
    api.get<FinalGrade[]>(`/subjects/${subjectId}/final-grades`)
};
