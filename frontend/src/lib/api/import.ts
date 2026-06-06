import { api } from './client';

export interface ImportStudentRow {
  full_name: string;
  email: string;
  code: string;
}

export interface ImportStructureRow {
  cut_number: number;
  cut_name: string;
  cut_weight: number;
  activity_name: string;
  activity_weight: number;
}

export interface ImportGradeRow {
  student_code: string;
  cut_number: number;
  activity_name: string;
  value: number;
}

export interface ImportPayload {
  students: ImportStudentRow[];
  structure: ImportStructureRow[];
  grades: ImportGradeRow[];
}

export interface ImportResult {
  students_created: number;
  students_enrolled: number;
  cuts_created: number;
  activities_created: number;
  grades_imported: number;
}

export const importApi = {
  submit: (subjectId: string, payload: ImportPayload) =>
    api.post<ImportResult>(`/subjects/${subjectId}/import`, payload)
};
