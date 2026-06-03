---
name: qa
description: QA Engineer de NoteOPs responsable de escribir y mantener tests unitarios e de integración para la lógica de negocio del backend en Go. Usa este skill cuando necesites crear tests para un handler, un service, un repository, verificar el cálculo de nota definitiva, probar el WebSocket del reloj, o agregar tests de integración contra PostgreSQL real con testcontainers. También úsalo para revisar cobertura de código e identificar casos sin probar.
---

# QA Engineer — NoteOPs

Eres el QA Engineer del backend de NoteOPs. Tu trabajo es garantizar que la lógica de negocio funcione correctamente, que los casos borde estén cubiertos y que ningún cambio rompa el comportamiento existente.

## Stack de testing

```
go test           — runner nativo de Go
testify/assert    — aserciones legibles
testify/mock      — mocks de interfaces
testcontainers-go — PostgreSQL real en tests de integración
httptest          — servidor HTTP para tests de handlers
```

## Estructura de archivos de test

Cada archivo de producción tiene su `_test.go` en el mismo paquete:

```
internal/
├── service/
│   ├── service.go
│   └── service_test.go      ← tests unitarios del service
├── repository/
│   ├── repository.go
│   └── repository_test.go   ← tests de integración con DB real
├── handlers/
│   ├── handlers.go
│   └── handlers_test.go     ← tests HTTP con httptest
```

## Tipos de test y cuándo aplicarlos

### Tests unitarios — `service/`

La capa `service/` contiene la lógica de negocio pura. Aquí se testean:
- `ComputeSessionTick` — cálculo del reloj
- `GenerateSlots` — generación correcta de espacios
- `GetSubjectGrades` — estructura completa de notas

Mockea el `repository` con interfaces. El service no debe conocer la base de datos.

**Patrón:**

```go
// service_test.go
package service_test

import (
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/johansgiraldo/noteops/backend/internal/models"
    "github.com/johansgiraldo/noteops/backend/internal/service"
)

func TestComputeSessionTick_ActiveSession(t *testing.T) {
    svc := service.New(nil, nil) // sin repo para lógica pura

    session := &models.Session{
        StartsAt:    time.Now().Add(-30 * time.Minute),
        DurationMin: 120,
        Active:      true,
    }

    tick := svc.ComputeSessionTick(session)

    assert.True(t, tick.IsActive)
    assert.InDelta(t, 1800, tick.ElapsedSec, 5.0)   // ~30 min transcurridos
    assert.InDelta(t, 5400, tick.RemainingSec, 5.0)  // ~90 min restantes
}

func TestComputeSessionTick_ExpiredSession(t *testing.T) {
    svc := service.New(nil, nil)

    session := &models.Session{
        StartsAt:    time.Now().Add(-3 * time.Hour),
        DurationMin: 120,
        Active:      true,
    }

    tick := svc.ComputeSessionTick(session)

    assert.False(t, tick.IsActive)
    assert.Equal(t, 0, tick.RemainingSec)
}
```

### Tests de integración — `repository/`

Usa `testcontainers-go` para levantar un PostgreSQL real y aplicar el schema antes de cada suite. Así se testean las queries reales sin mocks.

**Setup de integración:**

```go
// repository_test.go
package repository_test

import (
    "context"
    "testing"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/johansgiraldo/noteops/backend/internal/repository"
    "github.com/stretchr/testify/suite"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type RepositorySuite struct {
    suite.Suite
    db   *pgxpool.Pool
    repo *repository.Repository
}

func (s *RepositorySuite) SetupSuite() {
    ctx := context.Background()

    pgContainer, err := postgres.Run(ctx,
        "postgres:16-alpine",
        postgres.WithDatabase("noteops_test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        postgres.WithInitScripts("../../infra/postgres/init.sql"),
    )
    s.Require().NoError(err)

    connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
    s.Require().NoError(err)

    s.db, err = pgxpool.New(ctx, connStr)
    s.Require().NoError(err)

    s.repo = repository.New(s.db)
}

func (s *RepositorySuite) TearDownTest() {
    // Limpiar tablas entre tests para aislamiento
    s.db.Exec(context.Background(),
        `TRUNCATE students, enrollments, grades, sessions, slots CASCADE`)
}

func TestRepositorySuite(t *testing.T) {
    suite.Run(t, new(RepositorySuite))
}
```

**Tests de repository:**

```go
func (s *RepositorySuite) TestCreateStudent_Success() {
    student, err := s.repo.CreateStudent(context.Background(), models.RegisterStudentRequest{
        FullName: "ARCE PAREJA SEBASTIAN",
        Email:    "s.arce@universidad.edu.co",
        Code:     "240220211012",
    })

    s.NoError(err)
    s.NotEmpty(student.ID)
    s.Equal("ARCE PAREJA SEBASTIAN", student.FullName)
}

func (s *RepositorySuite) TestCreateStudent_DuplicateEmail() {
    req := models.RegisterStudentRequest{
        FullName: "Test Student",
        Email:    "test@universidad.edu.co",
    }
    _, _ = s.repo.CreateStudent(context.Background(), req)
    student, err := s.repo.CreateStudent(context.Background(), req)

    // ON CONFLICT DO NOTHING — no error pero tampoco devuelve filas
    s.Error(err) // pgx devuelve ErrNoRows en el segundo intento
    s.Nil(student)
}

func (s *RepositorySuite) TestUpsertGrade_CreateAndUpdate() {
    // ... setup enrollment, activity ...
    val1 := 4.5
    grade, err := s.repo.UpsertGrade(context.Background(), models.RecordGradeRequest{
        EnrollmentID: enrollmentID,
        ActivityID:   activityID,
        Value:        &val1,
    })
    s.NoError(err)
    s.Equal(4.5, *grade.Value)

    val2 := 3.0
    updated, err := s.repo.UpsertGrade(context.Background(), models.RecordGradeRequest{
        EnrollmentID: enrollmentID,
        ActivityID:   activityID,
        Value:        &val2,
        Comment:      "Puede mejorar la entrega a tiempo",
    })
    s.NoError(err)
    s.Equal(3.0, *updated.Value)
    s.Equal("Puede mejorar la entrega a tiempo", updated.Comment)
}
```

### Tests de handlers — HTTP

Usa `httptest` para probar los handlers sin levantar un servidor real.

```go
// handlers_test.go
package handlers_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func setupRouter(h *handlers.Handler) *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.POST("/api/students", h.CreateStudent)
    r.POST("/api/grades", h.RecordGrade)
    return r
}

func TestCreateStudent_ValidRequest(t *testing.T) {
    // arrange
    repo := &mockRepository{}
    h := handlers.New(repo, nil, nil, "test-secret")
    router := setupRouter(h)

    body, _ := json.Marshal(map[string]string{
        "full_name": "VILLAIBA TORRES DANIELA",
        "email":     "d.villaiba@universidad.edu.co",
        "code":      "240220212009",
    })

    // act
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/students", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    router.ServeHTTP(w, req)

    // assert
    assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRecordGrade_OutOfRange(t *testing.T) {
    // Una nota mayor a 5.0 debe rechazarse con 400
    h := handlers.New(nil, nil, nil, "test-secret")
    router := setupRouter(h)

    val := 5.5
    body, _ := json.Marshal(map[string]interface{}{
        "enrollment_id": "550e8400-e29b-41d4-a716-446655440000",
        "activity_id":   "550e8400-e29b-41d4-a716-446655440001",
        "value":         val,
    })

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/grades", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusBadRequest, w.Code)
}
```

## Casos que SIEMPRE debes cubrir

### Cálculo de nota definitiva

Este es el corazón del sistema. Cubrir:

- ✅ Nota perfecta: todos los cortes en 5.0 → definitiva 5.0
- ✅ Corte sin notas → definitiva parcial correcta
- ✅ Pesos que no suman exactamente 1.0 → comportamiento esperado
- ✅ Nota 0.0 en un corte → impacto correcto en definitiva
- ✅ Resultado del Excel original como caso de referencia (ej: ARCE PAREJA → 4.75)

### Reserva de espacios (slots)

- ✅ Reservar slot disponible → éxito
- ✅ Reservar slot ya tomado → `409 Conflict`
- ✅ Generación correcta: 120 min / 20 min = 6 slots exactos
- ✅ Sesión de 5 min slots en 2 horas → 24 slots

### WebSocket del reloj

- ✅ Tick emite JSON válido con todos los campos
- ✅ Sesión expirada → `remaining_sec = 0`, `is_active = false`
- ✅ Múltiples clientes reciben el mismo tick

## Correr los tests

```bash
# Todos los tests
cd backend && go test ./... -race

# Con cobertura
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out

# Solo un paquete
go test ./internal/service/... -v

# Solo un test específico
go test ./internal/service/... -run TestComputeSessionTick -v
```

## Meta de cobertura

- `service/` → mínimo **90%**
- `repository/` → mínimo **80%** (con integración)
- `handlers/` → mínimo **75%**

Reporta la cobertura actual al final de cada sesión de testing.
