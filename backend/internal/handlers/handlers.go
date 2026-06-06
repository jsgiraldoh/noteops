package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/johansgiraldo/noteops/backend/internal/middleware"
	"github.com/johansgiraldo/noteops/backend/internal/models"
	"github.com/johansgiraldo/noteops/backend/internal/repository"
	"github.com/johansgiraldo/noteops/backend/internal/service"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	repo      repository.Repo
	svc       *service.Service
	hub       *Hub
	jwtSecret string
}

func New(repo repository.Repo, svc *service.Service, hub *Hub, jwtSecret string) *Handler {
	return &Handler{repo: repo, svc: svc, hub: hub, jwtSecret: jwtSecret}
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

// POST /api/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": sanitizeBindError(err)})
		return
	}

	user, err := h.repo.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales incorrectas"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Credenciales incorrectas"})
		return
	}

	claims := &middleware.Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo generar el token de acceso"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": signed, "user": user})
}

// ─── Students ────────────────────────────────────────────────────────────────

// PATCH /api/students/:id
func (h *Handler) UpdateStudent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de estudiante inválido"})
		return
	}
	var req models.UpdateStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": sanitizeBindError(err)})
		return
	}
	student, err := h.repo.UpdateStudent(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusOK, student)
}

// POST /api/students
func (h *Handler) CreateStudent(c *gin.Context) {
	var req models.RegisterStudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": sanitizeBindError(err)})
		return
	}
	student, err := h.repo.CreateStudent(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusCreated, student)
}

// POST /api/subjects/:id/enroll
func (h *Handler) EnrollStudent(c *gin.Context) {
	subjectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de materia inválido"})
		return
	}
	var body struct {
		StudentID string `json:"student_id" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": sanitizeBindError(err)})
		return
	}
	studentID, _ := uuid.Parse(body.StudentID)
	enrollment, err := h.repo.EnrollStudent(c.Request.Context(), studentID, subjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusCreated, enrollment)
}

// GET /api/subjects/:id/students
func (h *Handler) GetStudentsBySubject(c *gin.Context) {
	subjectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de materia inválido"})
		return
	}
	students, err := h.repo.GetStudentsBySubject(c.Request.Context(), subjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusOK, students)
}

// ─── Subjects ────────────────────────────────────────────────────────────────

// GET /api/subjects
func (h *Handler) GetSubjects(c *gin.Context) {
	teacherID, _ := uuid.Parse(c.GetString("user_id"))
	subjects, err := h.repo.GetSubjectsByTeacher(c.Request.Context(), teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusOK, subjects)
}

// POST /api/subjects
func (h *Handler) CreateSubject(c *gin.Context) {
	var req models.CreateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": sanitizeBindError(err)})
		return
	}
	teacherID, _ := uuid.Parse(c.GetString("user_id"))
	subject, err := h.repo.CreateSubject(c.Request.Context(), req, teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusCreated, subject)
}

// PATCH /api/subjects/:id
func (h *Handler) UpdateSubject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de materia inválido"})
		return
	}
	var req models.UpdateSubjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": sanitizeBindError(err)})
		return
	}
	subject, err := h.repo.UpdateSubject(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusOK, subject)
}

// DELETE /api/subjects/:id
func (h *Handler) DeleteSubject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de materia inválido"})
		return
	}
	if err := h.repo.DeleteSubject(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true, "message": "Materia eliminada"})
}

// ─── Grades ──────────────────────────────────────────────────────────────────

// POST /api/grades
func (h *Handler) RecordGrade(c *gin.Context) {
	var req models.RecordGradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": sanitizeBindError(err)})
		return
	}
	grade, err := h.repo.UpsertGrade(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusOK, grade)
}

// PATCH /api/grades/:id/comment
func (h *Handler) UpdateComment(c *gin.Context) {
	gradeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de nota inválido"})
		return
	}
	var req models.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": sanitizeBindError(err)})
		return
	}
	if err := h.repo.UpdateComment(c.Request.Context(), gradeID, req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": true, "message": "Comentario actualizado"})
}

// GET /api/subjects/:id/grades
func (h *Handler) GetSubjectGrades(c *gin.Context) {
	subjectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de materia inválido"})
		return
	}
	result, err := h.svc.GetSubjectGrades(c.Request.Context(), subjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GET /api/subjects/:id/final-grades
func (h *Handler) GetFinalGrades(c *gin.Context) {
	subjectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de materia inválido"})
		return
	}
	grades, err := h.repo.GetFinalGradesBySubject(c.Request.Context(), subjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusOK, grades)
}

// ─── Sessions ────────────────────────────────────────────────────────────────

// POST /api/sessions
func (h *Handler) CreateSession(c *gin.Context) {
	var req models.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": sanitizeBindError(err)})
		return
	}
	session, err := h.repo.CreateSession(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	slots, err := h.svc.GenerateSlots(c.Request.Context(), session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sesión creada pero no se pudieron generar los espacios"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"session": session, "slots": slots})
}

// POST /api/sessions/:id/activate
func (h *Handler) ActivateSession(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de sesión inválido"})
		return
	}
	if err := h.repo.ActivateSession(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"activated": true, "message": "Sesión iniciada"})
}

// POST /api/sessions/:id/deactivate
func (h *Handler) DeactivateSession(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de sesión inválido"})
		return
	}
	if err := h.repo.DeactivateSession(c.Request.Context(), sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deactivated": true, "message": "Sesión finalizada"})
}

// GET /api/sessions/active?subject_id=uuid
func (h *Handler) GetActiveSession(c *gin.Context) {
	subjectID, err := uuid.Parse(c.Query("subject_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subject_id requerido y debe ser un UUID válido"})
		return
	}
	session, err := h.repo.GetActiveSessionBySubject(c.Request.Context(), subjectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No hay sesión activa para esta materia"})
		return
	}
	c.JSON(http.StatusOK, session)
}

// GET /api/sessions/:id/slots
func (h *Handler) GetSlots(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de sesión inválido"})
		return
	}
	slots, err := h.repo.GetSlotsBySession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusOK, slots)
}

// POST /api/sessions/:id/slots/:slotID/reserve
func (h *Handler) ReserveSlot(c *gin.Context) {
	slotID, err := uuid.Parse(c.Param("slotID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de espacio inválido"})
		return
	}
	var req models.ReserveSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": sanitizeBindError(err)})
		return
	}
	studentID, _ := uuid.Parse(req.StudentID)
	slot, err := h.repo.ReserveSlot(c.Request.Context(), slotID, studentID)
	if err != nil {
		// Ninguna fila actualizada: el espacio ya está tomado o la sesión ya no existe
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusConflict, gin.H{"error": "Este espacio ya fue reservado o la sesión ya no existe"})
			return
		}
		// Violación de llave foránea: el estudiante no existe en el sistema
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El estudiante seleccionado no existe o no está registrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusOK, slot)
}

// POST /api/subjects/:id/import
func (h *Handler) ImportSubjectData(c *gin.Context) {
	subjectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de materia inválido"})
		return
	}
	var req models.ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": sanitizeBindError(err)})
		return
	}
	result, err := h.repo.ImportSubjectData(c.Request.Context(), subjectID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GET /api/health
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
