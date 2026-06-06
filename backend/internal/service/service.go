package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johansgiraldo/noteops/backend/internal/models"
	"github.com/johansgiraldo/noteops/backend/internal/repository"
)

type Service struct {
	repo repository.Repo
	db   *pgxpool.Pool
}

func New(repo repository.Repo, db *pgxpool.Pool) *Service {
	return &Service{repo: repo, db: db}
}

// GenerateSlots crea los espacios de una sesión en la base de datos
func (s *Service) GenerateSlots(ctx context.Context, session *models.Session) ([]models.Slot, error) {
	count := session.DurationMin / session.SlotMin
	if count == 0 {
		count = 1
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var slots []models.Slot
	for i := 0; i < count; i++ {
		startsAt := session.StartsAt.Add(time.Duration(i*session.SlotMin) * time.Minute)
		var slot models.Slot
		err := tx.QueryRow(ctx,
			`INSERT INTO slots (session_id, number, starts_at, duration_min)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id, session_id, number, starts_at, duration_min, student_id, reserved_at`,
			session.ID, i+1, startsAt, session.SlotMin).
			Scan(&slot.ID, &slot.SessionID, &slot.Number, &slot.StartsAt,
				&slot.DurationMin, &slot.StudentID, &slot.ReservedAt)
		if err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return slots, nil
}

// GetSubjectGrades devuelve la estructura completa de notas de una materia:
// cortes → actividades → notas por estudiante
func (s *Service) GetSubjectGrades(ctx context.Context, subjectID uuid.UUID) (*SubjectGradesResult, error) {
	cuts, err := s.repo.GetCutsBySubject(ctx, subjectID)
	if err != nil {
		return nil, err
	}

	for i, cut := range cuts {
		activities, err := s.repo.GetActivitiesByCut(ctx, cut.ID)
		if err != nil {
			return nil, err
		}
		cuts[i].Activities = activities
	}

	students, err := s.repo.GetStudentsBySubject(ctx, subjectID)
	if err != nil {
		return nil, err
	}

	enrollments, err := s.repo.GetEnrollmentsBySubject(ctx, subjectID)
	if err != nil {
		return nil, err
	}

	grades, err := s.repo.GetGradesBySubject(ctx, subjectID)
	if err != nil {
		return nil, err
	}

	finalGrades, err := s.repo.GetFinalGradesBySubject(ctx, subjectID)
	if err != nil {
		return nil, err
	}

	return &SubjectGradesResult{
		Cuts:        cuts,
		Students:    students,
		Enrollments: enrollments,
		Grades:      grades,
		FinalGrades: finalGrades,
	}, nil
}

// ComputeSessionTick devuelve el estado actual del reloj de una sesión
func (s *Service) ComputeSessionTick(session *models.Session) models.SessionTick {
	now := time.Now()
	elapsed := int(now.Sub(session.StartsAt).Seconds())
	total := session.DurationMin * 60
	remaining := total - elapsed

	if remaining < 0 {
		remaining = 0
	}

	return models.SessionTick{
		SessionID:    session.ID.String(),
		ElapsedSec:   elapsed,
		RemainingSec: remaining,
		DurationMin:  session.DurationMin,
		IsActive:     session.Active && remaining > 0,
	}
}

// ─── Result types ─────────────────────────────────────────────────────────────

type SubjectGradesResult struct {
	Cuts        []models.Cut        `json:"cuts"`
	Students    []models.Student    `json:"students"`
	Enrollments []models.Enrollment `json:"enrollments"`
	Grades      []models.Grade      `json:"grades"`
	FinalGrades []models.FinalGrade `json:"final_grades"`
}
