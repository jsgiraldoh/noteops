package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/johansgiraldo/noteops/backend/internal/handlers"
	"github.com/johansgiraldo/noteops/backend/internal/middleware"
	"github.com/johansgiraldo/noteops/backend/internal/models"
	"github.com/johansgiraldo/noteops/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const testSecret = "test-secret-minimum-32-chars-long!!"

// ─── Mock repositorio ────────────────────────────────────────────────────────

type stubRepo struct {
	getUserByEmail      func(email string) (*models.User, error)
	getActiveSession    func(subjectID uuid.UUID) (*models.Session, error)
	getSubjectsByTeacher func() ([]models.Subject, error)
	reserveSlot         func(slotID, studentID uuid.UUID) (*models.Slot, error)
}

func (s *stubRepo) GetUserByEmail(_ context.Context, email string) (*models.User, error) {
	if s.getUserByEmail != nil {
		return s.getUserByEmail(email)
	}
	return nil, errors.New("not found")
}
func (s *stubRepo) GetActiveSessionBySubject(_ context.Context, id uuid.UUID) (*models.Session, error) {
	if s.getActiveSession != nil {
		return s.getActiveSession(id)
	}
	return nil, errors.New("not found")
}
func (s *stubRepo) GetSubjectsByTeacher(_ context.Context, _ uuid.UUID) ([]models.Subject, error) {
	if s.getSubjectsByTeacher != nil {
		return s.getSubjectsByTeacher()
	}
	return []models.Subject{}, nil
}
func (s *stubRepo) ReserveSlot(_ context.Context, slotID, studentID uuid.UUID) (*models.Slot, error) {
	if s.reserveSlot != nil {
		return s.reserveSlot(slotID, studentID)
	}
	return nil, errors.New("not found")
}

// Implementaciones vacías para satisfacer la interfaz completa
func (s *stubRepo) CreateStudent(_ context.Context, _ models.RegisterStudentRequest) (*models.Student, error) {
	return nil, nil
}
func (s *stubRepo) UpdateStudent(_ context.Context, _ uuid.UUID, _ models.UpdateStudentRequest) (*models.Student, error) {
	return nil, nil
}
func (s *stubRepo) GetStudentsBySubject(_ context.Context, _ uuid.UUID) ([]models.Student, error) {
	return []models.Student{}, nil
}
func (s *stubRepo) EnrollStudent(_ context.Context, _, _ uuid.UUID) (*models.Enrollment, error) {
	return nil, nil
}
func (s *stubRepo) GetEnrollmentsBySubject(_ context.Context, _ uuid.UUID) ([]models.Enrollment, error) {
	return []models.Enrollment{}, nil
}
func (s *stubRepo) CreateSubject(_ context.Context, _ models.CreateSubjectRequest, _ uuid.UUID) (*models.Subject, error) {
	return nil, nil
}
func (s *stubRepo) UpdateSubject(_ context.Context, _ uuid.UUID, _ models.UpdateSubjectRequest) (*models.Subject, error) {
	return nil, nil
}
func (s *stubRepo) DeleteSubject(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubRepo) GetSubjectByID(_ context.Context, _ uuid.UUID) (*models.Subject, error) {
	return nil, nil
}
func (s *stubRepo) GetCutsBySubject(_ context.Context, _ uuid.UUID) ([]models.Cut, error) {
	return []models.Cut{}, nil
}
func (s *stubRepo) GetActivitiesByCut(_ context.Context, _ uuid.UUID) ([]models.Activity, error) {
	return []models.Activity{}, nil
}
func (s *stubRepo) UpsertGrade(_ context.Context, _ models.RecordGradeRequest) (*models.Grade, error) {
	return nil, nil
}
func (s *stubRepo) UpdateComment(_ context.Context, _ uuid.UUID, _ string) error { return nil }
func (s *stubRepo) GetGradesBySubject(_ context.Context, _ uuid.UUID) ([]models.Grade, error) {
	return []models.Grade{}, nil
}
func (s *stubRepo) GetGradesByEnrollment(_ context.Context, _ uuid.UUID) ([]models.Grade, error) {
	return []models.Grade{}, nil
}
func (s *stubRepo) GetFinalGradesBySubject(_ context.Context, _ uuid.UUID) ([]models.FinalGrade, error) {
	return []models.FinalGrade{}, nil
}
func (s *stubRepo) CreateSession(_ context.Context, _ models.CreateSessionRequest) (*models.Session, error) {
	return nil, nil
}
func (s *stubRepo) GetSessionByID(_ context.Context, _ uuid.UUID) (*models.Session, error) {
	return nil, nil
}
func (s *stubRepo) ActivateSession(_ context.Context, _ uuid.UUID) error   { return nil }
func (s *stubRepo) DeactivateSession(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubRepo) GetSlotsBySession(_ context.Context, _ uuid.UUID) ([]models.Slot, error) {
	return []models.Slot{}, nil
}
func (s *stubRepo) ImportSubjectData(_ context.Context, _ uuid.UUID, _ models.ImportRequest) (*models.ImportResult, error) {
	return nil, nil
}

// Verifica en tiempo de compilación que stubRepo implementa la interfaz
var _ repository.Repo = (*stubRepo)(nil)

// ─── Router de test ──────────────────────────────────────────────────────────

func newTestRouter(repo repository.Repo) *gin.Engine {
	r := gin.New()
	h := handlers.New(repo, nil, nil, testSecret)

	r.GET("/api/health", h.Health)
	r.POST("/api/auth/login", h.Login)
	r.GET("/api/sessions/active", h.GetActiveSession)

	auth := r.Group("/api", middleware.Auth(testSecret))
	{
		auth.GET("/subjects", h.GetSubjects)
		auth.POST("/sessions/:id/slots/:slotID/reserve", h.ReserveSlot)
	}
	return r
}

func doPost(router *gin.Engine, path string, body any, token ...string) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	if len(token) > 0 {
		req.Header.Set("Authorization", "Bearer "+token[0])
	}
	router.ServeHTTP(w, req)
	return w
}

func doGet(router *gin.Engine, path string, token ...string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	if len(token) > 0 {
		req.Header.Set("Authorization", "Bearer "+token[0])
	}
	router.ServeHTTP(w, req)
	return w
}

// loginAndGetToken realiza un login real contra el router de test
// y devuelve el JWT para usarlo en tests de endpoints protegidos.
func loginAndGetToken(t *testing.T) string {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte("pwd"), bcrypt.MinCost)
	repo := &stubRepo{
		getUserByEmail: func(_ string) (*models.User, error) {
			return &models.User{
				ID:       uuid.New(),
				Email:    "test@test.com",
				Password: string(hash),
				Role:     "admin",
			}, nil
		},
	}
	r := newTestRouter(repo)
	w := doPost(r, "/api/auth/login", map[string]string{
		"email": "test@test.com", "password": "pwd",
	})
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	token, ok := resp["token"].(string)
	if !ok || token == "" {
		t.Fatal("no se pudo obtener token de test")
	}
	return token
}

// ─── GET /api/health ─────────────────────────────────────────────────────────

func TestHealth_Returns200WithStatusOK(t *testing.T) {
	r := newTestRouter(&stubRepo{})
	w := doGet(r, "/api/health")

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, quería %d", w.Code, http.StatusOK)
	}
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("status = %q, quería \"ok\"", body["status"])
	}
}

// ─── POST /api/auth/login ─────────────────────────────────────────────────────

func TestLogin_ValidCredentials_Returns200AndToken(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.MinCost)
	r := newTestRouter(&stubRepo{
		getUserByEmail: func(_ string) (*models.User, error) {
			return &models.User{
				ID: uuid.New(), Email: "admin@noteops.local",
				Password: string(hash), Role: "admin",
			}, nil
		},
	})

	w := doPost(r, "/api/auth/login", map[string]string{
		"email": "admin@noteops.local", "password": "admin123",
	})

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, quería %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}
	var body map[string]any
	json.NewDecoder(w.Body).Decode(&body)
	if body["token"] == nil || body["token"] == "" {
		t.Error("la respuesta debe contener el campo token")
	}
}

func TestLogin_WrongPassword_Returns401(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	r := newTestRouter(&stubRepo{
		getUserByEmail: func(_ string) (*models.User, error) {
			return &models.User{Password: string(hash)}, nil
		},
	})

	w := doPost(r, "/api/auth/login", map[string]string{
		"email": "admin@noteops.local", "password": "wrong",
	})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, quería %d", w.Code, http.StatusUnauthorized)
	}
}

func TestLogin_UnknownEmail_Returns401(t *testing.T) {
	r := newTestRouter(&stubRepo{
		getUserByEmail: func(_ string) (*models.User, error) {
			return nil, errors.New("no rows")
		},
	})

	w := doPost(r, "/api/auth/login", map[string]string{
		"email": "nobody@noteops.local", "password": "admin123",
	})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, quería %d", w.Code, http.StatusUnauthorized)
	}
}

func TestLogin_MissingFields_Returns400(t *testing.T) {
	r := newTestRouter(&stubRepo{})
	w := doPost(r, "/api/auth/login", map[string]string{"email": "only@email.com"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, quería %d", w.Code, http.StatusBadRequest)
	}
}

func TestLogin_ErrorMessageDoesNotLeakDetails(t *testing.T) {
	r := newTestRouter(&stubRepo{
		getUserByEmail: func(_ string) (*models.User, error) {
			return nil, errors.New("pq: relation \"users\" does not exist")
		},
	})

	w := doPost(r, "/api/auth/login", map[string]string{
		"email": "x@x.com", "password": "x",
	})

	body := w.Body.String()
	if strings.Contains(body, "relation") || strings.Contains(body, "pq:") || strings.Contains(body, "does not exist") {
		t.Errorf("respuesta filtra detalle interno: %s", body)
	}
}

// ─── Middleware de autenticación ──────────────────────────────────────────────

func TestProtectedEndpoint_WithoutToken_Returns401(t *testing.T) {
	r := newTestRouter(&stubRepo{})
	w := doGet(r, "/api/subjects")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("sin token = %d, quería %d", w.Code, http.StatusUnauthorized)
	}
}

func TestProtectedEndpoint_WithValidToken_Returns200(t *testing.T) {
	token := loginAndGetToken(t)
	r := newTestRouter(&stubRepo{})
	w := doGet(r, "/api/subjects", token)
	if w.Code != http.StatusOK {
		t.Errorf("con token válido = %d, quería %d", w.Code, http.StatusOK)
	}
}

func TestProtectedEndpoint_WithInvalidToken_Returns401(t *testing.T) {
	r := newTestRouter(&stubRepo{})
	w := doGet(r, "/api/subjects", "token.invalido.aqui")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("token inválido = %d, quería %d", w.Code, http.StatusUnauthorized)
	}
}

// ─── GET /api/sessions/active ─────────────────────────────────────────────────

func TestGetActiveSession_MissingSubjectID_Returns400(t *testing.T) {
	r := newTestRouter(&stubRepo{})
	w := doGet(r, "/api/sessions/active")
	if w.Code != http.StatusBadRequest {
		t.Errorf("sin subject_id = %d, quería %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetActiveSession_InvalidUUID_Returns400(t *testing.T) {
	r := newTestRouter(&stubRepo{})
	w := doGet(r, "/api/sessions/active?subject_id=not-a-uuid")
	if w.Code != http.StatusBadRequest {
		t.Errorf("UUID inválido = %d, quería %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetActiveSession_NoSession_Returns404(t *testing.T) {
	r := newTestRouter(&stubRepo{
		getActiveSession: func(_ uuid.UUID) (*models.Session, error) {
			return nil, errors.New("no rows")
		},
	})
	w := doGet(r, "/api/sessions/active?subject_id="+uuid.New().String())
	if w.Code != http.StatusNotFound {
		t.Errorf("sin sesión activa = %d, quería %d", w.Code, http.StatusNotFound)
	}
}

// ─── POST reserve slot ────────────────────────────────────────────────────────

func TestReserveSlot_InvalidSlotID_Returns400(t *testing.T) {
	token := loginAndGetToken(t)
	r := newTestRouter(&stubRepo{})
	sessionID := uuid.New().String()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost,
		"/api/sessions/"+sessionID+"/slots/not-a-uuid/reserve",
		bytes.NewBufferString(`{"student_id":"550e8400-e29b-41d4-a716-446655440000"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("slotID inválido = %d, quería %d", w.Code, http.StatusBadRequest)
	}
}

func TestReserveSlot_InvalidJSON_Returns400(t *testing.T) {
	token := loginAndGetToken(t)
	r := newTestRouter(&stubRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost,
		"/api/sessions/"+uuid.New().String()+"/slots/"+uuid.New().String()+"/reserve",
		bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("JSON inválido = %d, quería %d", w.Code, http.StatusBadRequest)
	}
}
