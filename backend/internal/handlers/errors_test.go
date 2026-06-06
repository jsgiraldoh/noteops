package handlers

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ─── safeError ───────────────────────────────────────────────────────────────

func TestSafeError_Nil(t *testing.T) {
	if got := safeError(nil); got != "" {
		t.Errorf("safeError(nil) = %q, quería %q", got, "")
	}
}

func TestSafeError_NoRows(t *testing.T) {
	got := safeError(pgx.ErrNoRows)
	want := "Registro no encontrado"
	if got != want {
		t.Errorf("safeError(pgx.ErrNoRows) = %q, quería %q", got, want)
	}
}

func TestSafeError_UniqueViolation(t *testing.T) {
	err := &pgconn.PgError{Code: "23505"}
	got := safeError(err)
	want := "Ya existe un registro con ese dato"
	if got != want {
		t.Errorf("safeError(23505) = %q, quería %q", got, want)
	}
}

func TestSafeError_ForeignKeyViolation(t *testing.T) {
	err := &pgconn.PgError{Code: "23503"}
	got := safeError(err)
	want := "No se puede eliminar porque tiene datos relacionados"
	if got != want {
		t.Errorf("safeError(23503) = %q, quería %q", got, want)
	}
}

func TestSafeError_CheckViolation(t *testing.T) {
	err := &pgconn.PgError{Code: "23514"}
	got := safeError(err)
	want := "El valor está fuera del rango permitido (0–5)"
	if got != want {
		t.Errorf("safeError(23514) = %q, quería %q", got, want)
	}
}

func TestSafeError_NotNullViolation(t *testing.T) {
	err := &pgconn.PgError{Code: "23502"}
	got := safeError(err)
	want := "Faltan campos obligatorios"
	if got != want {
		t.Errorf("safeError(23502) = %q, quería %q", got, want)
	}
}

func TestSafeError_InvalidTextRepresentation(t *testing.T) {
	err := &pgconn.PgError{Code: "22P02"}
	got := safeError(err)
	want := "Formato de dato inválido"
	if got != want {
		t.Errorf("safeError(22P02) = %q, quería %q", got, want)
	}
}

func TestSafeError_SerializationFailure(t *testing.T) {
	err := &pgconn.PgError{Code: "40001"}
	got := safeError(err)
	want := "Conflicto de concurrencia, intenta de nuevo"
	if got != want {
		t.Errorf("safeError(40001) = %q, quería %q", got, want)
	}
}

func TestSafeError_UnknownPgError(t *testing.T) {
	err := &pgconn.PgError{Code: "99999"}
	got := safeError(err)
	want := "Error al procesar la solicitud"
	if got != want {
		t.Errorf("safeError(PgError desconocido) = %q, quería %q", got, want)
	}
}

func TestSafeError_GenericError(t *testing.T) {
	err := errors.New("algún error interno con stack trace y detalles")
	got := safeError(err)
	want := "Error interno del servidor"
	if got != want {
		t.Errorf("safeError(error genérico) = %q, quería %q", got, want)
		t.Error("ERROR: se está filtrando información interna al cliente")
	}
}

// ─── sanitizeBindError ───────────────────────────────────────────────────────

func TestSanitizeBindError_Required(t *testing.T) {
	err := errors.New("Field validation for 'Name' failed on the 'required' tag")
	got := sanitizeBindError(err)
	want := "Hay campos obligatorios sin completar"
	if got != want {
		t.Errorf("sanitizeBindError(required) = %q, quería %q", got, want)
	}
}

func TestSanitizeBindError_Email(t *testing.T) {
	err := errors.New("Field validation for 'Email' failed on the 'email' tag")
	got := sanitizeBindError(err)
	want := "El correo electrónico no tiene un formato válido"
	if got != want {
		t.Errorf("sanitizeBindError(email) = %q, quería %q", got, want)
	}
}

func TestSanitizeBindError_UUID(t *testing.T) {
	err := errors.New("Field validation for 'SubjectID' failed on the 'uuid' tag")
	got := sanitizeBindError(err)
	want := "Uno o más identificadores tienen formato inválido"
	if got != want {
		t.Errorf("sanitizeBindError(uuid) = %q, quería %q", got, want)
	}
}

func TestSanitizeBindError_EOF(t *testing.T) {
	err := errors.New("EOF")
	got := sanitizeBindError(err)
	want := "El cuerpo de la solicitud tiene formato inválido"
	if got != want {
		t.Errorf("sanitizeBindError(EOF) = %q, quería %q", got, want)
	}
}

func TestSanitizeBindError_Unknown(t *testing.T) {
	err := errors.New("algún error de validación no contemplado")
	got := sanitizeBindError(err)
	want := "Datos de entrada inválidos"
	if got != want {
		t.Errorf("sanitizeBindError(desconocido) = %q, quería %q", got, want)
	}
}
