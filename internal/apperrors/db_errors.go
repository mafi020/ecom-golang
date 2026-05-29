// internal/apperrors/db_errors.go

package apperrors

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func HandleUniqueViolation(err error, constraints map[string]string) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		field, ok := constraints[pgErr.ConstraintName]
		if !ok {
			field = "record"
		}
		return &ConflictError{
			Errors: map[string]string{field: field + " already exists"},
		}
	}
	return err
}
