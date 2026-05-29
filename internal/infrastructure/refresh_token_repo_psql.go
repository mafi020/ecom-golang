package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/entity"
)

type PostgresRefreshTokenRepository struct {
	db *sql.DB
}

func NewPostgresRefreshTokenRepository(db *sql.DB) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

func (r *PostgresRefreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO refresh_tokens(user_id, token, expires_at) VALUES($1, $2, $3)",
		token.UserID, token.Token, token.ExpiresAt,
	)
	return fmt.Errorf("failed to create refresh token: %w", err)

}

func (r *PostgresRefreshTokenRepository) GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	query := `SELECT id, user_id, token, expires_at FROM refresh_tokens WHERE token = $1`
	row := r.db.QueryRowContext(ctx, query, token)

	var rt entity.RefreshToken
	err := row.Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "Refresh Token"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &rt, nil
}

func (r *PostgresRefreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE token = $1", token)

	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return &apperrors.NotFoundError{Resource: "Refresh Token"}
	}

	return nil
}

func (r *PostgresRefreshTokenRepository) DeleteAllByUserID(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE user_id = $1", userID)
	return err
}
