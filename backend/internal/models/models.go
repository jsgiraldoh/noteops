package models

import (
	"time"

	"github.com/google/uuid"
)

// ─── User ────────────────────────────────────────────────────────────────────

type User struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"` // teacher | admin
	CreatedAt time.Time `json:"created_at"`
}

// ─── Subject ─────────────────────────────────────────────────────────────────

type Subject struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Period    string    `json:"period"`
	GroupName string    `json:"group_name"`
	Faculty   string    `json:"faculty"`
	TeacherID uuid.UUID `json:"teacher_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ─── Cut ─────────────────────────────────────────────────────────────────────

type Cut struct {
	ID         uuid.UUID  `json:"id"`
	SubjectID  uuid.UUID  `json:"subject_id"`
	Number     int        `json:"number"`
	Name       string     `json:"name"`
	Weight     float64    `json:"weight"`
	Activities []Activity `json:"activities,omitempty"`
}

// ─── Activity ────────────────────────────────────────────────────────────────

type Activity struct {
	ID          uuid.UUID  `json:"id"`
	CutID       uuid.UUID  `json:"cut_id"`
	Name        string     `json:"name"`
	Weight      float64    `json:"weight"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

// ─── Student ─────────────────────────────────────────────────────────────────

type Student struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

// ─── Enrollment ──────────────────────────────────────────────────────────────

type Enrollment struct {
	ID        uuid.UUID `json:"id"`
	StudentID uuid.UUID `json:"student_id"`
	SubjectID uuid.UUID `json:"subject_id"`
}

// ─── Grade ───────────────────────────────────────────────────────────────────

type Grade struct {
	ID           uuid.UUID `json:"id"`
	EnrollmentID uuid.UUID `json:"enrollment_id"`
	ActivityID   uuid.UUID `json:"activity_id"`
	Value        *float64  `json:"value"`
	Comment      string    `json:"comment"`
	RecordedAt   time.Time `json:"recorded_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ─── Session ─────────────────────────────────────────────────────────────────

type Session struct {
	ID          uuid.UUID `json:"id"`
	SubjectID   uuid.UUID `json:"subject_id"`
	StartsAt    time.Time `json:"starts_at"`
	DurationMin int       `json:"duration_min"`
	SlotMin     int       `json:"slot_min"`
	Room        string    `json:"room"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}

// ─── Slot ────────────────────────────────────────────────────────────────────

type Slot struct {
	ID          uuid.UUID  `json:"id"`
	SessionID   uuid.UUID  `json:"session_id"`
	Number      int        `json:"number"`
	StartsAt    time.Time  `json:"starts_at"`
	DurationMin int        `json:"duration_min"`
	StudentID   *uuid.UUID `json:"student_id,omitempty"`
	ReservedAt  *time.Time `json:"reserved_at,omitempty"`
}

// ─── DTOs ────────────────────────────────────────────────────────────────────

type CreateSubjectRequest struct {
	Name      string `json:"name"       binding:"required"`
	Period    string `json:"period"     binding:"required"`
	GroupName string `json:"group_name"`
	Faculty   string `json:"faculty"`
}

type UpdateSubjectRequest struct {
	Name      string `json:"name"       binding:"required"`
	Period    string `json:"period"     binding:"required"`
	GroupName string `json:"group_name"`
	Faculty   string `json:"faculty"`
}

type UpdateStudentRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email"     binding:"required,email"`
	Code     string `json:"code"`
}

type RegisterStudentRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email"     binding:"required,email"`
	Code     string `json:"code"`
}

type RecordGradeRequest struct {
	EnrollmentID string   `json:"enrollment_id" binding:"required,uuid"`
	ActivityID   string   `json:"activity_id"   binding:"required,uuid"`
	Value        *float64 `json:"value"         binding:"required,min=0,max=5"`
	Comment      string   `json:"comment"`
}

type UpdateCommentRequest struct {
	Comment string `json:"comment" binding:"required"`
}

type CreateSessionRequest struct {
	SubjectID   string `json:"subject_id"    binding:"required,uuid"`
	StartsAt    string `json:"starts_at"     binding:"required"`
	DurationMin int    `json:"duration_min"  binding:"required,min=5"`
	SlotMin     int    `json:"slot_min"      binding:"required,min=5"`
	Room        string `json:"room"`
}

type ReserveSlotRequest struct {
	StudentID string `json:"student_id" binding:"required,uuid"`
}

type ImportStudentRow struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Code     string `json:"code"`
}

type ImportStructureRow struct {
	CutNumber      int     `json:"cut_number"`
	CutName        string  `json:"cut_name"`
	CutWeight      float64 `json:"cut_weight"`
	ActivityName   string  `json:"activity_name"`
	ActivityWeight float64 `json:"activity_weight"`
}

type ImportGradeRow struct {
	StudentCode  string  `json:"student_code"`
	CutNumber    int     `json:"cut_number"`
	ActivityName string  `json:"activity_name"`
	Value        float64 `json:"value"`
}

type ImportRequest struct {
	Students  []ImportStudentRow  `json:"students"`
	Structure []ImportStructureRow `json:"structure"`
	Grades    []ImportGradeRow    `json:"grades"`
}

type ImportResult struct {
	StudentsCreated   int `json:"students_created"`
	StudentsEnrolled  int `json:"students_enrolled"`
	CutsCreated       int `json:"cuts_created"`
	ActivitiesCreated int `json:"activities_created"`
	GradesImported    int `json:"grades_imported"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ─── Responses ───────────────────────────────────────────────────────────────

type FinalGrade struct {
	EnrollmentID string  `json:"enrollment_id"`
	StudentID    string  `json:"student_id"`
	SubjectID    string  `json:"subject_id"`
	FinalGrade   float64 `json:"final_grade"`
}

type SessionTick struct {
	SessionID    string `json:"session_id"`
	ElapsedSec   int    `json:"elapsed_sec"`
	RemainingSec int    `json:"remaining_sec"`
	DurationMin  int    `json:"duration_min"`
	IsActive     bool   `json:"is_active"`
}
