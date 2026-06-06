//go:build integration

// Tests de integración contra un PostgreSQL real levantado con testcontainers.
// Se ejecutan solo con la build tag `integration`:
//
//	go test -tags=integration ./internal/repository/...
//
// Requieren un daemon de Docker accesible.
package repository_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johansgiraldo/noteops/backend/internal/models"
	"github.com/johansgiraldo/noteops/backend/internal/repository"
	"github.com/johansgiraldo/noteops/backend/internal/service"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RepositorySuite struct {
	suite.Suite
	container *postgres.PostgresContainer
	db        *pgxpool.Pool
	repo      *repository.Repository
	svc       *service.Service
	adminID   uuid.UUID
}

// initScriptPath resuelve la ruta absoluta a infra/postgres/init.sql
// relativa a este archivo de test, sin depender del working directory.
func initScriptPath() string {
	_, thisFile, _, _ := runtime.Caller(0)
	// thisFile = .../backend/internal/repository/repository_integration_test.go
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "infra", "postgres", "init.sql")
}

func (s *RepositorySuite) SetupSuite() {
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("noteops_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts(initScriptPath()),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	s.Require().NoError(err)
	s.container = container

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	s.Require().NoError(err)

	s.db, err = pgxpool.New(ctx, connStr)
	s.Require().NoError(err)
	s.Require().NoError(s.db.Ping(ctx))

	s.repo = repository.New(s.db)
	s.svc = service.New(s.repo, s.db)

	// El admin lo crea init.sql; lo usamos como teacher_id en las materias.
	admin, err := s.repo.GetUserByEmail(ctx, "admin@noteops.local")
	s.Require().NoError(err)
	s.adminID = admin.ID
}

func (s *RepositorySuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.container != nil {
		_ = s.container.Terminate(context.Background())
	}
}

// TearDownTest limpia los datos entre tests, preservando el usuario admin.
func (s *RepositorySuite) TearDownTest() {
	_, err := s.db.Exec(context.Background(),
		`TRUNCATE subjects, students RESTART IDENTITY CASCADE`)
	s.Require().NoError(err)
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (s *RepositorySuite) ctx() context.Context { return context.Background() }

func (s *RepositorySuite) newSubject(name string) *models.Subject {
	subj, err := s.repo.CreateSubject(s.ctx(), models.CreateSubjectRequest{
		Name:      name,
		Period:    "2025-1",
		GroupName: "A",
		Faculty:   "Ingeniería",
	}, s.adminID)
	s.Require().NoError(err)
	return subj
}

// ─── Users ───────────────────────────────────────────────────────────────────

func (s *RepositorySuite) TestGetUserByEmail_AdminExists() {
	admin, err := s.repo.GetUserByEmail(s.ctx(), "admin@noteops.local")
	s.NoError(err)
	s.Equal("admin", admin.Role)
	s.NotEmpty(admin.Password) // hash bcrypt presente
}

func (s *RepositorySuite) TestGetUserByEmail_NotFound() {
	_, err := s.repo.GetUserByEmail(s.ctx(), "nadie@noteops.local")
	s.Error(err)
}

// ─── Students ────────────────────────────────────────────────────────────────

func (s *RepositorySuite) TestCreateStudent_Success() {
	st, err := s.repo.CreateStudent(s.ctx(), models.RegisterStudentRequest{
		FullName: "ARCE PAREJA SEBASTIAN",
		Email:    "240220211012@noteops.edu",
		Code:     "240220211012",
	})
	s.NoError(err)
	s.NotEqual(uuid.Nil, st.ID)
	s.Equal("ARCE PAREJA SEBASTIAN", st.FullName)
}

func (s *RepositorySuite) TestCreateStudent_DuplicateEmail() {
	req := models.RegisterStudentRequest{
		FullName: "Test", Email: "dup@noteops.edu", Code: "111",
	}
	_, err := s.repo.CreateStudent(s.ctx(), req)
	s.Require().NoError(err)

	// ON CONFLICT DO NOTHING → no devuelve fila → ErrNoRows
	_, err = s.repo.CreateStudent(s.ctx(), req)
	s.Error(err)
}

func (s *RepositorySuite) TestUpdateStudent_ChangesFields() {
	st, _ := s.repo.CreateStudent(s.ctx(), models.RegisterStudentRequest{
		FullName: "Nombre Viejo", Email: "upd@noteops.edu", Code: "222",
	})
	updated, err := s.repo.UpdateStudent(s.ctx(), st.ID, models.UpdateStudentRequest{
		FullName: "Nombre Nuevo", Email: "upd@noteops.edu", Code: "999",
	})
	s.NoError(err)
	s.Equal("Nombre Nuevo", updated.FullName)
	s.Equal("999", updated.Code)
}

// ─── Enrollment ──────────────────────────────────────────────────────────────

func (s *RepositorySuite) TestEnrollStudent_AndIdempotency() {
	subj := s.newSubject("Sistemas Operativos")
	st, _ := s.repo.CreateStudent(s.ctx(), models.RegisterStudentRequest{
		FullName: "Estudiante", Email: "enr@noteops.edu", Code: "333",
	})

	enr, err := s.repo.EnrollStudent(s.ctx(), st.ID, subj.ID)
	s.NoError(err)
	s.NotEqual(uuid.Nil, enr.ID)

	// Segundo enroll → ON CONFLICT DO NOTHING → ErrNoRows
	_, err = s.repo.EnrollStudent(s.ctx(), st.ID, subj.ID)
	s.Error(err)

	// El estudiante aparece en la materia
	students, err := s.repo.GetStudentsBySubject(s.ctx(), subj.ID)
	s.NoError(err)
	s.Len(students, 1)
}

// ─── Subjects ────────────────────────────────────────────────────────────────

func (s *RepositorySuite) TestCreateSubject_AndGetByID() {
	subj := s.newSubject("Contenedores Docker")
	got, err := s.repo.GetSubjectByID(s.ctx(), subj.ID)
	s.NoError(err)
	s.Equal("Contenedores Docker", got.Name)
	s.Equal(s.adminID, got.TeacherID)
}

func (s *RepositorySuite) TestDeleteSubject_CascadesEnrollments() {
	subj := s.newSubject("Materia Temporal")
	st, _ := s.repo.CreateStudent(s.ctx(), models.RegisterStudentRequest{
		FullName: "X", Email: "casc@noteops.edu", Code: "444",
	})
	_, _ = s.repo.EnrollStudent(s.ctx(), st.ID, subj.ID)

	err := s.repo.DeleteSubject(s.ctx(), subj.ID)
	s.NoError(err)

	_, err = s.repo.GetSubjectByID(s.ctx(), subj.ID)
	s.Error(err) // ya no existe

	// La inscripción se eliminó en cascada; el estudiante sigue existiendo
	enrolls, err := s.repo.GetEnrollmentsBySubject(s.ctx(), subj.ID)
	s.NoError(err)
	s.Len(enrolls, 0)
}

// ─── Sessions & Slots ────────────────────────────────────────────────────────

func (s *RepositorySuite) TestGenerateSlots_CorrectCount() {
	subj := s.newSubject("Materia Sesión")
	sess, err := s.repo.CreateSession(s.ctx(), models.CreateSessionRequest{
		SubjectID:   subj.ID.String(),
		StartsAt:    time.Now().Format(time.RFC3339),
		DurationMin: 120,
		SlotMin:     20,
		Room:        "Sala 201",
	})
	s.Require().NoError(err)

	slots, err := s.svc.GenerateSlots(s.ctx(), sess)
	s.NoError(err)
	s.Len(slots, 6) // 120 / 20 = 6
	s.Equal(1, slots[0].Number)
	s.Equal(6, slots[5].Number)
}

func (s *RepositorySuite) TestReserveSlot_SuccessThenConflict() {
	subj := s.newSubject("Materia Reserva")
	st, _ := s.repo.CreateStudent(s.ctx(), models.RegisterStudentRequest{
		FullName: "Reservante", Email: "res@noteops.edu", Code: "555",
	})
	sess, _ := s.repo.CreateSession(s.ctx(), models.CreateSessionRequest{
		SubjectID:   subj.ID.String(),
		StartsAt:    time.Now().Format(time.RFC3339),
		DurationMin: 120,
		SlotMin:     20,
	})
	slots, _ := s.svc.GenerateSlots(s.ctx(), sess)

	// Primera reserva → éxito
	reserved, err := s.repo.ReserveSlot(s.ctx(), slots[0].ID, st.ID)
	s.NoError(err)
	s.NotNil(reserved.StudentID)
	s.Equal(st.ID, *reserved.StudentID)

	// Segunda reserva del mismo slot → WHERE student_id IS NULL no matchea → ErrNoRows
	_, err = s.repo.ReserveSlot(s.ctx(), slots[0].ID, st.ID)
	s.Error(err)
}

func (s *RepositorySuite) TestDeactivateSession_DeletesSessionAndSlots() {
	subj := s.newSubject("Materia Borrado")
	sess, _ := s.repo.CreateSession(s.ctx(), models.CreateSessionRequest{
		SubjectID:   subj.ID.String(),
		StartsAt:    time.Now().Format(time.RFC3339),
		DurationMin: 60,
		SlotMin:     20,
	})
	_, _ = s.svc.GenerateSlots(s.ctx(), sess)

	err := s.repo.DeactivateSession(s.ctx(), sess.ID)
	s.NoError(err)

	// La sesión ya no existe
	_, err = s.repo.GetSessionByID(s.ctx(), sess.ID)
	s.Error(err)

	// Los slots se eliminaron en cascada
	slots, err := s.repo.GetSlotsBySession(s.ctx(), sess.ID)
	s.NoError(err)
	s.Len(slots, 0)
}

// ─── Cálculo de nota definitiva (vista student_final_grades) ─────────────────

func (s *RepositorySuite) TestFinalGrade_PerfectScore() {
	subj := s.newSubject("Materia Perfecta")
	_, err := s.repo.ImportSubjectData(s.ctx(), subj.ID, models.ImportRequest{
		Students: []models.ImportStudentRow{
			{FullName: "Estudiante Perfecto", Email: "perfecto@noteops.edu", Code: "1000"},
		},
		Structure: []models.ImportStructureRow{
			{CutNumber: 1, CutName: "Único", CutWeight: 1.0, ActivityName: "Final", ActivityWeight: 1.0},
		},
		Grades: []models.ImportGradeRow{
			{StudentCode: "1000", CutNumber: 1, ActivityName: "Final", Value: 5.0},
		},
	})
	s.Require().NoError(err)

	finals, err := s.repo.GetFinalGradesBySubject(s.ctx(), subj.ID)
	s.NoError(err)
	s.Require().Len(finals, 1)
	s.InDelta(5.0, finals[0].FinalGrade, 0.001)
}

// TestFinalGrade_ArcePareja reproduce el caso real de la planilla del Excel:
// ARCE PAREJA SEBASTIAN en Sistemas Operativos → nota definitiva 4.75.
func (s *RepositorySuite) TestFinalGrade_ArcePareja() {
	subj := s.newSubject("Sistemas Operativos")

	req := models.ImportRequest{
		Students: []models.ImportStudentRow{
			{FullName: "ARCE PAREJA SEBASTIAN", Email: "240220211012@noteops.edu", Code: "240220211012"},
		},
		Structure: []models.ImportStructureRow{
			// Corte 1 — peso 0.3
			{CutNumber: 1, CutName: "Primer Corte", CutWeight: 0.3, ActivityName: "N1", ActivityWeight: 0.17},
			{CutNumber: 1, CutName: "Primer Corte", CutWeight: 0.3, ActivityName: "N2", ActivityWeight: 0.17},
			{CutNumber: 1, CutName: "Primer Corte", CutWeight: 0.3, ActivityName: "N3", ActivityWeight: 0.16},
			{CutNumber: 1, CutName: "Primer Corte", CutWeight: 0.3, ActivityName: "PC", ActivityWeight: 0.5},
			// Corte 2 — peso 0.3
			{CutNumber: 2, CutName: "Segundo Corte", CutWeight: 0.3, ActivityName: "N1", ActivityWeight: 0.1},
			{CutNumber: 2, CutName: "Segundo Corte", CutWeight: 0.3, ActivityName: "N2", ActivityWeight: 0.1},
			{CutNumber: 2, CutName: "Segundo Corte", CutWeight: 0.3, ActivityName: "N3", ActivityWeight: 0.1},
			{CutNumber: 2, CutName: "Segundo Corte", CutWeight: 0.3, ActivityName: "N4", ActivityWeight: 0.1},
			{CutNumber: 2, CutName: "Segundo Corte", CutWeight: 0.3, ActivityName: "N5", ActivityWeight: 0.1},
			{CutNumber: 2, CutName: "Segundo Corte", CutWeight: 0.3, ActivityName: "PC", ActivityWeight: 0.5},
			// Corte 3 — peso 0.4
			{CutNumber: 3, CutName: "Tercer Corte", CutWeight: 0.4, ActivityName: "N1", ActivityWeight: 0.13},
			{CutNumber: 3, CutName: "Tercer Corte", CutWeight: 0.4, ActivityName: "N2", ActivityWeight: 0.13},
			{CutNumber: 3, CutName: "Tercer Corte", CutWeight: 0.4, ActivityName: "N3", ActivityWeight: 0.12},
			{CutNumber: 3, CutName: "Tercer Corte", CutWeight: 0.4, ActivityName: "N4", ActivityWeight: 0.12},
			{CutNumber: 3, CutName: "Tercer Corte", CutWeight: 0.4, ActivityName: "PC", ActivityWeight: 0.5},
		},
		Grades: []models.ImportGradeRow{
			// Corte 1 — todo 5 → DFC1 = 5.0
			{StudentCode: "240220211012", CutNumber: 1, ActivityName: "N1", Value: 5},
			{StudentCode: "240220211012", CutNumber: 1, ActivityName: "N2", Value: 5},
			{StudentCode: "240220211012", CutNumber: 1, ActivityName: "N3", Value: 5},
			{StudentCode: "240220211012", CutNumber: 1, ActivityName: "PC", Value: 5},
			// Corte 2 — todo 5 → DFC2 = 5.0
			{StudentCode: "240220211012", CutNumber: 2, ActivityName: "N1", Value: 5},
			{StudentCode: "240220211012", CutNumber: 2, ActivityName: "N2", Value: 5},
			{StudentCode: "240220211012", CutNumber: 2, ActivityName: "N3", Value: 5},
			{StudentCode: "240220211012", CutNumber: 2, ActivityName: "N4", Value: 5},
			{StudentCode: "240220211012", CutNumber: 2, ActivityName: "N5", Value: 5},
			{StudentCode: "240220211012", CutNumber: 2, ActivityName: "PC", Value: 5},
			// Corte 3 → DFC3 = 4.375
			{StudentCode: "240220211012", CutNumber: 3, ActivityName: "N1", Value: 5},
			{StudentCode: "240220211012", CutNumber: 3, ActivityName: "N2", Value: 2.5},
			{StudentCode: "240220211012", CutNumber: 3, ActivityName: "N3", Value: 3.75},
			{StudentCode: "240220211012", CutNumber: 3, ActivityName: "N4", Value: 3.75},
			{StudentCode: "240220211012", CutNumber: 3, ActivityName: "PC", Value: 5},
		},
	}

	res, err := s.repo.ImportSubjectData(s.ctx(), subj.ID, req)
	s.Require().NoError(err)
	s.Equal(15, res.GradesImported)
	s.Equal(3, res.CutsCreated)

	finals, err := s.repo.GetFinalGradesBySubject(s.ctx(), subj.ID)
	s.NoError(err)
	s.Require().Len(finals, 1)
	// DFC1·0.3 + DFC2·0.3 + DFC3·0.4 = 1.5 + 1.5 + 1.75 = 4.75
	s.InDelta(4.75, finals[0].FinalGrade, 0.001)
}

// ─── Import end-to-end ───────────────────────────────────────────────────────

func (s *RepositorySuite) TestImportSubjectData_StructureOnly() {
	subj := s.newSubject("Materia Solo Estructura")
	res, err := s.repo.ImportSubjectData(s.ctx(), subj.ID, models.ImportRequest{
		Students: []models.ImportStudentRow{
			{FullName: "A B", Email: "ab@noteops.edu", Code: "2001"},
			{FullName: "C D", Email: "cd@noteops.edu", Code: "2002"},
		},
		Structure: []models.ImportStructureRow{
			{CutNumber: 1, CutName: "C1", CutWeight: 0.5, ActivityName: "N1", ActivityWeight: 1.0},
			{CutNumber: 2, CutName: "C2", CutWeight: 0.5, ActivityName: "N1", ActivityWeight: 1.0},
		},
		Grades: nil, // sin notas
	})
	s.NoError(err)
	s.Equal(2, res.StudentsCreated)
	s.Equal(2, res.StudentsEnrolled)
	s.Equal(2, res.CutsCreated)
	s.Equal(2, res.ActivitiesCreated)
	s.Equal(0, res.GradesImported)

	cuts, err := s.repo.GetCutsBySubject(s.ctx(), subj.ID)
	s.NoError(err)
	s.Len(cuts, 2)
}
