package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/entity"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
        INSERT INTO users(name, email, password, role) 
        VALUES($1, $2, $3, $4) 
        RETURNING id, name, email, created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Name,
		user.Email,
		user.Password,
		user.Role,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return apperrors.HandleUniqueViolation(err, map[string]string{
			"users_email_key": "email",
		})
	}

	return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	query := `
		SELECT id, name, email, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var user entity.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "user"}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user:: %w", err)
	}

	return &user, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, name, email, password, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	row := r.db.QueryRowContext(ctx, query, email)

	var user entity.User
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "user"}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *PostgresUserRepository) GetAll(ctx context.Context, params entity.GetUsersParams) ([]entity.User, int, error) {
	// Base query
	baseQuery := `FROM users WHERE 1=1`
	args := []any{}
	argIndex := 1

	// Search
	if params.Search != "" {
		baseQuery += fmt.Sprintf(" AND (name ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex+1)
		search := "%" + params.Search + "%"
		args = append(args, search, search)
		argIndex += 2
	}

	// Filter by role
	if params.Role != "" {
		baseQuery += fmt.Sprintf(" AND role = $%d", argIndex)
		args = append(args, params.Role)
		argIndex++
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to fetch users:: %w", err)
	}

	// Allowed sort columns to prevent SQL injection
	allowedSortColumns := map[string]bool{
		"created_at": true,
		"name":       true,
		"email":      true,
	}
	if !allowedSortColumns[params.SortBy] {
		params.SortBy = "created_at"
	}
	if params.SortOrder != "asc" {
		params.SortOrder = "desc"
	}

	// Main query with sort and pagination
	mainQuery := fmt.Sprintf(
		"SELECT id, name, email, role, created_at, updated_at %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		baseQuery, params.SortBy, params.SortOrder, argIndex, argIndex+1,
	)

	args = append(args, params.Limit, params.Offset())

	rows, err := r.db.QueryContext(ctx, mainQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch users:: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)

	if err != nil {
		return fmt.Errorf("failed to delete user:: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected:: %w", err)
	}

	if rows == 0 {
		return &apperrors.NotFoundError{Resource: "user"}
	}

	return nil
}
