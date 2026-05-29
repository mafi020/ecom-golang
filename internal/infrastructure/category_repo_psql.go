package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/entity"
)

type PostgresCategoryRepository struct {
	db *sql.DB
}

func NewPostgresCategoryRepository(db *sql.DB) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{db: db}
}

func (r *PostgresCategoryRepository) CreateCategory(ctx context.Context, category *entity.Category) error {
	query := `INSERT INTO categories (name, slug, parent_id) 
	VALUES ($1, $2, $3)
	Returning id, name, slug, parent_id, created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx,
		query,
		category.Name,
		category.Slug,
		category.ParentID,
	).Scan(&category.ID, &category.Name, &category.Slug, &category.ParentID, &category.CreatedAt, &category.UpdatedAt)

	if err != nil {
		return apperrors.HandleUniqueViolation(err, map[string]string{
			"categories_name_key": "name",
			"categories_slug_key": "slug",
		})
	}

	return nil
}

func (r *PostgresCategoryRepository) GetCategoryByID(ctx context.Context, id int64) (*entity.Category, error) {
	query := `SELECT id, name, slug, parent_id, created_at, updated_at 
	FROM categories 
	WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)

	category := &entity.Category{}

	err := row.Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.ParentID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "category"}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}

func (r *PostgresCategoryRepository) GetCategoryByIDWithProducts(ctx context.Context, id int64) (*entity.Category, error) {
	query := `
		SELECT c.id, c.name, c.slug, c.parent_id, c.created_at, c.updated_at, p.id, p.name, p.description, p.price 
		FROM categories c 
		LEFT JOIN products p ON c.id = p.category_id 
		WHERE c.id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	category := &entity.Category{}

	err := row.Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.ParentID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "category"}
	}

	if err != nil {
		return nil, err
	}

	return category, nil
}

func (r *PostgresCategoryRepository) GetAllCategories(ctx context.Context, params entity.GetCategoriesParams) ([]entity.Category, int, error) {
	baseQuery := `FROM categories WHERE 1=1`
	args := []any{}
	argIndex := 1

	if params.Search != "" {
		baseQuery += fmt.Sprintf(" AND (name ILIKE $%d OR slug ILIKE $%d)", argIndex, argIndex+1)
		search := "%" + params.Search + "%"
		args = append(args, search, search)
		argIndex += 2
	}

	if params.ParentID > 0 {
		baseQuery += fmt.Sprintf(" AND parent_id = $%d", argIndex)
		args = append(args, params.ParentID)
		argIndex++
	}

	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count categories: %w", err)
	}

	allowedSortColumns := map[string]bool{
		"created_at": true,
		"name":       true,
		"slug":       true,
	}

	if !allowedSortColumns[params.SortBy] {
		params.SortBy = "created_at"
	}
	if params.SortOrder != "asc" {
		params.SortOrder = "desc"
	}

	mainQuery := fmt.Sprintf(
		"SELECT id, name, slug, parent_id, created_at, updated_at %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		baseQuery, params.SortBy, params.SortOrder, argIndex, argIndex+1,
	)
	args = append(args, params.Limit, params.Offset())

	rows, err := r.db.QueryContext(ctx, mainQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch categories: %w", err)
	}
	defer rows.Close()

	var categories []entity.Category
	for rows.Next() {
		var category entity.Category
		var parentID sql.NullInt64

		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&parentID,
			&category.CreatedAt,
			&category.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan category: %w", err)
		}

		if parentID.Valid {
			category.ParentID = &parentID.Int64
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return categories, total, nil
}

// func (r *PostgresCategoryRepository) GetAllCategories(ctx context.Context) ([]*entity.Category, error) {
// 	query := `SELECT id, name, slug, parent_id, created_at, updated_at
// 	FROM categories`

// 	rows, err := r.db.QueryContext(ctx, query)

// 	if err != nil {
// 		return nil, err
// 	}

// 	defer rows.Close()

// 	var categories []*entity.Category

// 	for rows.Next() {
// 		category := &entity.Category{}
// 		var parentID sql.NullInt64 // use NullInt64 to safely scan nullable column
// 		err := rows.Scan(
// 			&category.ID,
// 			&category.Name,
// 			&category.Slug,
// 			&parentID,
// 			&category.CreatedAt,
// 			&category.UpdatedAt,
// 		)

// 		if err != nil {
// 			return nil, err
// 		}

// 		// Only assign if the value is not NULL
// 		if parentID.Valid {
// 			category.ParentID = &parentID.Int64
// 		}

// 		categories = append(categories, category)
// 	}
// 	return categories, nil
// }

func (r *PostgresCategoryRepository) UpdateCategory(ctx context.Context, category *entity.Category) error {
	query := `
		UPDATE categories
		SET name = $1, slug = $2, parent_id = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, name, slug, parent_id, created_at, updated_at`

	err := r.db.QueryRowContext(
		ctx,
		query,
		category.Name,
		category.Slug,
		category.ParentID,
		category.ID,
	).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.ParentID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &apperrors.NotFoundError{Resource: "category"}
		}

		err = apperrors.HandleUniqueViolation(err, map[string]string{
			"categories_name_unique": "name",
			"categories_slug_key":    "slug",
		})

		var conflictErr *apperrors.ConflictError
		if errors.As(err, &conflictErr) {
			return err
		}
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

func (r *PostgresCategoryRepository) DeleteCategory(ctx context.Context, id int64) error {
	query := `DELETE FROM categories WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)

	if err != nil {
		return fmt.Errorf("failed to delete category:: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected:: %w", err)
	}

	if rows == 0 {
		return &apperrors.NotFoundError{Resource: "category"}
	}
	return nil
}

func (r *PostgresCategoryRepository) IsCycle(ctx context.Context, categoryID int64, parentID int64) (bool, error) {
	query := `
		WITH RECURSIVE category_tree AS (
			SELECT id, parent_id FROM categories WHERE id = $1
			UNION ALL
			SELECT c.id, c.parent_id FROM categories c
			JOIN category_tree ct ON ct.parent_id = c.id
		)
		SELECT EXISTS (SELECT 1 FROM category_tree WHERE id = $2)`

	var isCycle bool
	err := r.db.QueryRowContext(ctx, query, parentID, categoryID).Scan(&isCycle)
	return isCycle, err
}
