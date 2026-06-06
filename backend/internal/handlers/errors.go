package handlers

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// safeError convierte errores internos en mensajes seguros para el frontend.
// Nunca expone detalles de SQL, nombres de tablas ni stack traces.
func safeError(err error) string {
	if err == nil {
		return ""
	}

	// Sin filas — registro no encontrado
	if errors.Is(err, pgx.ErrNoRows) {
		return "Registro no encontrado"
	}

	// Errores de PostgreSQL por código
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return "Ya existe un registro con ese dato"
		case "23503":
			return "No se puede eliminar porque tiene datos relacionados"
		case "23514":
			return "El valor está fuera del rango permitido (0–5)"
		case "23502":
			return "Faltan campos obligatorios"
		case "22P02", "22003":
			return "Formato de dato inválido"
		case "40001":
			return "Conflicto de concurrencia, intenta de nuevo"
		}
		// Cualquier otro error de PostgreSQL — no revelar detalles
		return "Error al procesar la solicitud"
	}

	return "Error interno del servidor"
}

// sanitizeBindError convierte los mensajes de validación de gin en español.
func sanitizeBindError(err error) string {
	msg := err.Error()
	// Mensajes de binding comunes
	switch {
	case strings.Contains(msg, "required"):
		return "Hay campos obligatorios sin completar"
	case strings.Contains(msg, "email"):
		return "El correo electrónico no tiene un formato válido"
	case strings.Contains(msg, "min="):
		return "Uno o más valores están por debajo del mínimo permitido"
	case strings.Contains(msg, "max="):
		return "Uno o más valores superan el máximo permitido"
	case strings.Contains(msg, "uuid"):
		return "Uno o más identificadores tienen formato inválido"
	case strings.Contains(msg, "EOF"), strings.Contains(msg, "cannot unmarshal"):
		return "El cuerpo de la solicitud tiene formato inválido"
	}
	return "Datos de entrada inválidos"
}
