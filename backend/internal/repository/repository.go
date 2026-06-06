package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johansgiraldo/noteops/backend/internal/models"
)

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ─── Users ───────────────────────────────────────────────────────────────────

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, full_name, email, password, role, created_at
		 FROM users WHERE email = $1`, email).
		Scan(&u.ID, &u.FullName, &u.Email, &u.Password, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// ─── Students ────────────────────────────────────────────────────────────────

func (r *Repository) CreateStudent(ctx context.Context, req models.RegisterStudentRequest) (*models.Student, error) {
	s := &models.Student{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO students (full_name, email, code)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (email) DO NOTHING
		 RETURNING id, full_name, email, code, created_at`,
		req.FullName, req.Email, req.Code).
		Scan(&s.ID, &s.FullName, &s.Email, &s.Code, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) GetStudentsBySubject(ctx context.Context, subjectID uuid.UUID) ([]models.Student, error) {
	rows, err := r.db.Query(ctx,
		`SELECT s.id, s.full_name, s.email, s.code, s.created_at
		 FROM students s
		 JOIN enrollments e ON e.student_id = s.id
		 WHERE e.subject_id = $1
		 ORDER BY s.full_name`, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(&s.ID, &s.FullName, &s.Email, &s.Code, &s.CreatedAt); err != nil {
			return nil, err
		}
		students = append(students, s)
	}
	return students, nil
}

func (r *Repository) EnrollStudent(ctx context.Context, studentID, subjectID uuid.UUID) (*models.Enrollment, error) {
	e := &models.Enrollment{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO enrollments (student_id, subject_id)
		 VALUES ($1, $2)
		 ON CONFLICT (student_id, subject_id) DO NOTHING
		 RETURNING id, student_id, subject_id`,
		studentID, subjectID).
		Scan(&e.ID, &e.StudentID, &e.SubjectID)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (r *Repository) GetEnrollmentsBySubject(ctx context.Context, subjectID uuid.UUID) ([]models.Enrollment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, student_id, subject_id FROM enrollments WHERE subject_id = $1`, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enrollments []models.Enrollment
	for rows.Next() {
		var e models.Enrollment
		if err := rows.Scan(&e.ID, &e.StudentID, &e.SubjectID); err != nil {
			return nil, err
		}
		enrollments = append(enrollments, e)
	}
	return enrollments, nil
}

func (r *Repository) GetGradesBySubject(ctx context.Context, subjectID uuid.UUID) ([]models.Grade, error) {
	rows, err := r.db.Query(ctx,
		`SELECT g.id, g.enrollment_id, g.activity_id, g.value, COALESCE(g.comment,''), g.recorded_at, g.updated_at
		 FROM grades g
		 JOIN enrollments e ON e.id = g.enrollment_id
		 WHERE e.subject_id = $1`, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grades []models.Grade
	for rows.Next() {
		var g models.Grade
		if err := rows.Scan(&g.ID, &g.EnrollmentID, &g.ActivityID, &g.Value,
			&g.Comment, &g.RecordedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		grades = append(grades, g)
	}
	return grades, nil
}

// ─── Subjects ────────────────────────────────────────────────────────────────

func (r *Repository) GetSubjectsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]models.Subject, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, period, group_name, faculty, teacher_id, created_at
		 FROM subjects WHERE teacher_id = $1 ORDER BY created_at DESC`, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subjects []models.Subject
	for rows.Next() {
		var s models.Subject
		if err := rows.Scan(&s.ID, &s.Name, &s.Period, &s.GroupName,
			&s.Faculty, &s.TeacherID, &s.CreatedAt); err != nil {
			return nil, err
		}
		subjects = append(subjects, s)
	}
	return subjects, nil
}

func (r *Repository) GetSubjectByID(ctx context.Context, id uuid.UUID) (*models.Subject, error) {
	s := &models.Subject{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, period, group_name, faculty, teacher_id, created_at
		 FROM subjects WHERE id = $1`, id).
		Scan(&s.ID, &s.Name, &s.Period, &s.GroupName, &s.Faculty, &s.TeacherID, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// ─── Cuts & Activities ───────────────────────────────────────────────────────

func (r *Repository) GetCutsBySubject(ctx context.Context, subjectID uuid.UUID) ([]models.Cut, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, subject_id, number, name, weight
		 FROM cuts WHERE subject_id = $1 ORDER BY number`, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cuts []models.Cut
	for rows.Next() {
		var c models.Cut
		if err := rows.Scan(&c.ID, &c.SubjectID, &c.Number, &c.Name, &c.Weight); err != nil {
			return nil, err
		}
		cuts = append(cuts, c)
	}
	return cuts, nil
}

func (r *Repository) GetActivitiesByCut(ctx context.Context, cutID uuid.UUID) ([]models.Activity, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, cut_id, name, weight, scheduled_at
		 FROM activities WHERE cut_id = $1 ORDER BY name`, cutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var acts []models.Activity
	for rows.Next() {
		var a models.Activity
		if err := rows.Scan(&a.ID, &a.CutID, &a.Name, &a.Weight, &a.ScheduledAt); err != nil {
			return nil, err
		}
		acts = append(acts, a)
	}
	return acts, nil
}

// ─── Grades ──────────────────────────────────────────────────────────────────

func (r *Repository) UpsertGrade(ctx context.Context, req models.RecordGradeRequest) (*models.Grade, error) {
	enrollmentID, _ := uuid.Parse(req.EnrollmentID)
	activityID, _ := uuid.Parse(req.ActivityID)

	g := &models.Grade{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO grades (enrollment_id, activity_id, value, comment)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (enrollment_id, activity_id) DO UPDATE
		   SET value = EXCLUDED.value,
		       comment = EXCLUDED.comment,
		       updated_at = NOW()
		 RETURNING id, enrollment_id, activity_id, value, COALESCE(comment,''), recorded_at, updated_at`,
		enrollmentID, activityID, req.Value, req.Comment).
		Scan(&g.ID, &g.EnrollmentID, &g.ActivityID, &g.Value,
			&g.Comment, &g.RecordedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (r *Repository) UpdateComment(ctx context.Context, gradeID uuid.UUID, comment string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE grades SET comment = $1, updated_at = NOW() WHERE id = $2`,
		comment, gradeID)
	return err
}

func (r *Repository) GetGradesByEnrollment(ctx context.Context, enrollmentID uuid.UUID) ([]models.Grade, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, enrollment_id, activity_id, value, COALESCE(comment,''), recorded_at, updated_at
		 FROM grades WHERE enrollment_id = $1`, enrollmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grades []models.Grade
	for rows.Next() {
		var g models.Grade
		if err := rows.Scan(&g.ID, &g.EnrollmentID, &g.ActivityID, &g.Value,
			&g.Comment, &g.RecordedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		grades = append(grades, g)
	}
	return grades, nil
}

func (r *Repository) GetFinalGradesBySubject(ctx context.Context, subjectID uuid.UUID) ([]models.FinalGrade, error) {
	rows, err := r.db.Query(ctx,
		`SELECT enrollment_id, student_id, subject_id, final_grade
		 FROM student_final_grades
		 WHERE subject_id = $1`, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.FinalGrade
	for rows.Next() {
		var fg models.FinalGrade
		if err := rows.Scan(&fg.EnrollmentID, &fg.StudentID,
			&fg.SubjectID, &fg.FinalGrade); err != nil {
			return nil, err
		}
		results = append(results, fg)
	}
	return results, nil
}

// ─── Sessions ────────────────────────────────────────────────────────────────

func (r *Repository) CreateSession(ctx context.Context, req models.CreateSessionRequest) (*models.Session, error) {
	subjectID, _ := uuid.Parse(req.SubjectID)
	startsAt, _ := time.Parse(time.RFC3339, req.StartsAt)

	s := &models.Session{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO sessions (subject_id, starts_at, duration_min, slot_min, room)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, subject_id, starts_at, duration_min, slot_min, room, active, created_at`,
		subjectID, startsAt, req.DurationMin, req.SlotMin, req.Room).
		Scan(&s.ID, &s.SubjectID, &s.StartsAt, &s.DurationMin,
			&s.SlotMin, &s.Room, &s.Active, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	s := &models.Session{}
	err := r.db.QueryRow(ctx,
		`SELECT id, subject_id, starts_at, duration_min, slot_min, room, active, created_at
		 FROM sessions WHERE id = $1`, id).
		Scan(&s.ID, &s.SubjectID, &s.StartsAt, &s.DurationMin,
			&s.SlotMin, &s.Room, &s.Active, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) ActivateSession(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE sessions SET active = true WHERE id = $1`, id)
	return err
}

func (r *Repository) DeactivateSession(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM sessions WHERE id = $1`, id)
	return err
}

func (r *Repository) GetActiveSessionBySubject(ctx context.Context, subjectID uuid.UUID) (*models.Session, error) {
	s := &models.Session{}
	err := r.db.QueryRow(ctx,
		`SELECT id, subject_id, starts_at, duration_min, slot_min, room, active, created_at
		 FROM sessions WHERE subject_id = $1 AND active = true
		 ORDER BY created_at DESC LIMIT 1`, subjectID).
		Scan(&s.ID, &s.SubjectID, &s.StartsAt, &s.DurationMin,
			&s.SlotMin, &s.Room, &s.Active, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repository) GetSlotsBySession(ctx context.Context, sessionID uuid.UUID) ([]models.Slot, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, session_id, number, starts_at, duration_min, student_id, reserved_at
		 FROM slots WHERE session_id = $1 ORDER BY number`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []models.Slot
	for rows.Next() {
		var s models.Slot
		if err := rows.Scan(&s.ID, &s.SessionID, &s.Number, &s.StartsAt,
			&s.DurationMin, &s.StudentID, &s.ReservedAt); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, nil
}

func (r *Repository) ReserveSlot(ctx context.Context, slotID, studentID uuid.UUID) (*models.Slot, error) {
	s := &models.Slot{}
	err := r.db.QueryRow(ctx,
		`UPDATE slots
		 SET student_id = $1, reserved_at = NOW()
		 WHERE id = $2 AND student_id IS NULL
		 RETURNING id, session_id, number, starts_at, duration_min, student_id, reserved_at`,
		studentID, slotID).
		Scan(&s.ID, &s.SessionID, &s.Number, &s.StartsAt,
			&s.DurationMin, &s.StudentID, &s.ReservedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}
