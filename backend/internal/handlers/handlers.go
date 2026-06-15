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

// enrollStudentBody es el cuerpo de la petición para matricular un estudiante.
type enrollStudentBody struct {
	StudentID string `json:"student_id" binding:"required,uuid"`
}

// ─── Auth ─────────────────────────────────────────────────────────────────────

// Login godoc
// @Summary      Autenticar usuario
// @Description  Valida credenciales y devuelve un JWT válido por 24 horas.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      models.LoginRequest   true  "Correo y contraseña"
// @Success      200          {object}  models.LoginResponse
// @Failure      400          {object}  models.ErrorResponse
// @Failure      401          {object}  models.ErrorResponse
// @Failure      500          {object}  models.ErrorResponse
// @Router       /auth/login [post]
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

// UpdateStudent godoc
// @Summary      Actualizar estudiante
// @Tags         students
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                       true  "ID del estudiante (UUID)"
// @Param        student  body      models.UpdateStudentRequest  true  "Datos del estudiante"
// @Success      200      {object}  models.Student
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /students/{id} [patch]
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

// CreateStudent godoc
// @Summary      Crear estudiante
// @Tags         students
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        student  body      models.RegisterStudentRequest  true  "Datos del estudiante"
// @Success      201      {object}  models.Student
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /students [post]
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

// EnrollStudent godoc
// @Summary      Matricular estudiante en una materia
// @Tags         subjects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                                true  "ID de la materia (UUID)"
// @Param        body  body      handlers.enrollStudentBody            true  "ID del estudiante"
// @Success      201   {object}  models.Enrollment
// @Failure      400   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Router       /subjects/{id}/enroll [post]
func (h *Handler) EnrollStudent(c *gin.Context) {
	subjectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de materia inválido"})
		return
	}
	var body enrollStudentBody
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

// GetStudentsBySubject godoc
// @Summary      Listar estudiantes de una materia
// @Tags         subjects
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string           true  "ID de la materia (UUID)"
// @Success      200  {array}   models.Student
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /subjects/{id}/students [get]
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

// GetSubjects godoc
// @Summary      Listar materias del docente autenticado
// @Tags         subjects
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   models.Subject
// @Failure      500  {object}  models.ErrorResponse
// @Router       /subjects [get]
func (h *Handler) GetSubjects(c *gin.Context) {
	teacherID, _ := uuid.Parse(c.GetString("user_id"))
	subjects, err := h.repo.GetSubjectsByTeacher(c.Request.Context(), teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": safeError(err)})
		return
	}
	c.JSON(http.StatusOK, subjects)
}

// CreateSubject godoc
// @Summary      Crear materia
// @Tags         subjects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        subject  body      models.CreateSubjectRequest  true  "Datos de la materia"
// @Success      201      {object}  models.Subject
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /subjects [post]
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

// UpdateSubject godoc
// @Summary      Actualizar materia
// @Tags         subjects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                       true  "ID de la materia (UUID)"
// @Param        subject  body      models.UpdateSubjectRequest  true  "Datos de la materia"
// @Success      200      {object}  models.Subject
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /subjects/{id} [patch]
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

// DeleteSubject godoc
// @Summary      Eliminar materia
// @Tags         subjects
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string                  true  "ID de la materia (UUID)"
// @Success      200  {object}  models.MessageResponse
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /subjects/{id} [delete]
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

// RecordGrade godoc
// @Summary      Registrar o actualizar una nota
// @Description  Crea o actualiza (upsert) la nota de una actividad para una matrícula.
// @Tags         grades
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        grade  body      models.RecordGradeRequest  true  "Datos de la nota (valor entre 0 y 5)"
// @Success      200    {object}  models.Grade
// @Failure      400    {object}  models.ErrorResponse
// @Failure      500    {object}  models.ErrorResponse
// @Router       /grades [post]
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

// UpdateComment godoc
// @Summary      Actualizar el comentario de una nota
// @Tags         grades
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                       true  "ID de la nota (UUID)"
// @Param        comment  body      models.UpdateCommentRequest  true  "Nuevo comentario"
// @Success      200      {object}  models.MessageResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /grades/{id}/comment [patch]
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

// GetSubjectGrades godoc
// @Summary      Obtener la planilla completa de notas de una materia
// @Description  Devuelve cortes, estudiantes, matrículas, notas y notas definitivas calculadas.
// @Tags         grades
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string                        true  "ID de la materia (UUID)"
// @Success      200  {object}  service.SubjectGradesResult
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /subjects/{id}/grades [get]
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

// GetFinalGrades godoc
// @Summary      Obtener las notas definitivas de una materia
// @Tags         grades
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string              true  "ID de la materia (UUID)"
// @Success      200  {array}   models.FinalGrade
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /subjects/{id}/final-grades [get]
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

// CreateSession godoc
// @Summary      Crear sesión de clase con sus espacios
// @Description  Crea una sesión y genera automáticamente los espacios (slots) según la duración.
// @Tags         sessions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        session  body      models.CreateSessionRequest  true  "Datos de la sesión"
// @Success      201      {object}  models.CreateSessionResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /sessions [post]
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

// ActivateSession godoc
// @Summary      Iniciar una sesión de clase
// @Tags         sessions
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string                  true  "ID de la sesión (UUID)"
// @Success      200  {object}  models.MessageResponse
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /sessions/{id}/activate [post]
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

// DeactivateSession godoc
// @Summary      Finalizar una sesión de clase
// @Tags         sessions
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string                  true  "ID de la sesión (UUID)"
// @Success      200  {object}  models.MessageResponse
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /sessions/{id}/deactivate [post]
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

// GetActiveSession godoc
// @Summary      Obtener la sesión activa de una materia
// @Tags         sessions
// @Produce      json
// @Param        subject_id  query     string           true  "ID de la materia (UUID)"
// @Success      200         {object}  models.Session
// @Failure      400         {object}  models.ErrorResponse
// @Failure      404         {object}  models.ErrorResponse
// @Router       /sessions/active [get]
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

// GetSlots godoc
// @Summary      Listar los espacios (slots) de una sesión
// @Tags         slots
// @Produce      json
// @Param        id   path      string           true  "ID de la sesión (UUID)"
// @Success      200  {array}   models.Slot
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /sessions/{id}/slots [get]
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

// ReserveSlot godoc
// @Summary      Reservar un espacio de una sesión
// @Description  Reserva un turno para un estudiante. Devuelve 409 si el espacio ya está tomado.
// @Tags         slots
// @Accept       json
// @Produce      json
// @Param        id      path      string                     true  "ID de la sesión (UUID)"
// @Param        slotID  path      string                     true  "ID del espacio (UUID)"
// @Param        body    body      models.ReserveSlotRequest  true  "ID del estudiante"
// @Success      200     {object}  models.Slot
// @Failure      400     {object}  models.ErrorResponse
// @Failure      409     {object}  models.ErrorResponse
// @Failure      500     {object}  models.ErrorResponse
// @Router       /sessions/{id}/slots/{slotID}/reserve [post]
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

// ImportSubjectData godoc
// @Summary      Importar datos masivos de una materia
// @Description  Importa estudiantes, estructura de evaluación y notas en una sola operación.
// @Tags         subjects
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                true  "ID de la materia (UUID)"
// @Param        data  body      models.ImportRequest  true  "Datos a importar"
// @Success      200   {object}  models.ImportResult
// @Failure      400   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Router       /subjects/{id}/import [post]
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

// Health godoc
// @Summary      Healthcheck del servicio
// @Tags         health
// @Produce      json
// @Success      200  {object}  models.MessageResponse
// @Router       /health [get]
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
