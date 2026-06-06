package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/johansgiraldo/noteops/backend/internal/models"
)

// Repo es la interfaz que expone todas las operaciones de base de datos.
// Permite mockear el repositorio en tests unitarios sin necesidad de una BD real.
type Repo interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)

	CreateStudent(ctx context.Context, req models.RegisterStudentRequest) (*models.Student, error)
	UpdateStudent(ctx context.Context, id uuid.UUID, req models.UpdateStudentRequest) (*models.Student, error)
	GetStudentsBySubject(ctx context.Context, subjectID uuid.UUID) ([]models.Student, error)
	EnrollStudent(ctx context.Context, studentID, subjectID uuid.UUID) (*models.Enrollment, error)
	GetEnrollmentsBySubject(ctx context.Context, subjectID uuid.UUID) ([]models.Enrollment, error)

	CreateSubject(ctx context.Context, req models.CreateSubjectRequest, teacherID uuid.UUID) (*models.Subject, error)
	UpdateSubject(ctx context.Context, id uuid.UUID, req models.UpdateSubjectRequest) (*models.Subject, error)
	DeleteSubject(ctx context.Context, id uuid.UUID) error
	GetSubjectsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]models.Subject, error)
	GetSubjectByID(ctx context.Context, id uuid.UUID) (*models.Subject, error)

	GetCutsBySubject(ctx context.Context, subjectID uuid.UUID) ([]models.Cut, error)
	GetActivitiesByCut(ctx context.Context, cutID uuid.UUID) ([]models.Activity, error)

	UpsertGrade(ctx context.Context, req models.RecordGradeRequest) (*models.Grade, error)
	UpdateComment(ctx context.Context, gradeID uuid.UUID, comment string) error
	GetGradesBySubject(ctx context.Context, subjectID uuid.UUID) ([]models.Grade, error)
	GetGradesByEnrollment(ctx context.Context, enrollmentID uuid.UUID) ([]models.Grade, error)
	GetFinalGradesBySubject(ctx context.Context, subjectID uuid.UUID) ([]models.FinalGrade, error)

	CreateSession(ctx context.Context, req models.CreateSessionRequest) (*models.Session, error)
	GetSessionByID(ctx context.Context, id uuid.UUID) (*models.Session, error)
	ActivateSession(ctx context.Context, id uuid.UUID) error
	DeactivateSession(ctx context.Context, id uuid.UUID) error
	GetActiveSessionBySubject(ctx context.Context, subjectID uuid.UUID) (*models.Session, error)
	GetSlotsBySession(ctx context.Context, sessionID uuid.UUID) ([]models.Slot, error)
	ReserveSlot(ctx context.Context, slotID, studentID uuid.UUID) (*models.Slot, error)

	ImportSubjectData(ctx context.Context, subjectID uuid.UUID, req models.ImportRequest) (*models.ImportResult, error)
}
