package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/entity"
)

type PostgresProductRepository struct {
	db *sql.DB
}

func NewPostgresProductRepository(db *sql.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (r *PostgresProductRepository) Create(ctx context.Context, product *entity.Product) error {
	query := `
		INSERT INTO products (name, description, price, stock, category_id, created_at, updated_at)		
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, description, price, stock, category_id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		product.Stock,
		product.CategoryID,
		time.Now(),
		time.Now(),
	).Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.CategoryID, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return err
	}
	return fmt.Errorf("failed to create product %w", err)
}

func (r *PostgresProductRepository) List(ctx context.Context, params entity.GetProductsParams) ([]entity.Product, int, error) {
	// 1. Base query
	baseQuery := `FROM products WHERE 1=1`
	args := []any{}
	argIndex := 1

	// 2. Search by name or description
	if params.Search != "" {
		baseQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex+1)
		search := "%" + params.Search + "%"
		args = append(args, search, search)
		argIndex += 2
	}

	if params.CategoryID > 0 {
		baseQuery += fmt.Sprintf(" AND category_id = $%d", argIndex)
		args = append(args, params.CategoryID)
		argIndex++
	}

	// 4. Get total count
	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to fetch products %w", err)
	}

	// 5. Sorting Safety (Whitelisting columns)
	allowedSortColumns := map[string]bool{
		"created_at": true,
		"name":       true,
		"price":      true,
		"stock":      true,
	}

	if !allowedSortColumns[params.SortBy] {
		params.SortBy = "created_at"
	}
	if params.SortOrder != "asc" {
		params.SortOrder = "desc"
	}

	// 6. Main query with pagination
	mainQuery := fmt.Sprintf(
		"SELECT id, name, description, price, stock, category_id, created_at, updated_at %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		baseQuery, params.SortBy, params.SortOrder, argIndex, argIndex+1,
	)

	args = append(args, params.Limit, params.Offset())

	rows, err := r.db.QueryContext(ctx, mainQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch products %w", err)
	}
	defer rows.Close()

	var products []entity.Product
	for rows.Next() {
		var p entity.Product
		var desc sql.NullString
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&desc,
			&p.Price,
			&p.Stock,
			&p.CategoryID,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if desc.Valid {
			p.Description = &desc.String
		}

		products = append(products, p)
	}

	return products, total, nil
}

func (r *PostgresProductRepository) GetByID(ctx context.Context, id int64) (*entity.Product, error) {
	query := `
		SELECT id, name, description, price, stock, category_id, created_at, updated_at	
		FROM products
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	product := &entity.Product{}

	err := row.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.CategoryID,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "Product"}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to scan products %w", err)
	}

	return product, nil
}

func (r *PostgresProductRepository) Update(ctx context.Context, id int64, input *entity.UpdateProductInput) (*entity.Product, error) {
	query := `
    UPDATE products
	SET
		name        = COALESCE($1, name),
		description = COALESCE($2, description),
		price       = COALESCE($3, price),
		stock       = COALESCE($4, stock),
		category_id = COALESCE($5, category_id),
		updated_at  = NOW()
	WHERE id = $6
	RETURNING id, name, description, price, stock, category_id, created_at, updated_at
	`
	product := &entity.Product{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		input.Name,
		input.Description,
		input.Price,
		input.Stock,
		input.CategoryID,
		id,
	).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Stock,
		&product.CategoryID,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "Product"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

func (r *PostgresProductRepository) BatchUpdate(ctx context.Context, updates map[int64]*entity.UpdateProductInput) error {
	if len(updates) == 0 {
		return nil
	}

	// Builds:
	// UPDATE products SET
	//     name        = CASE WHEN id = $1 THEN COALESCE($2,  name)        ELSE name        END,
	//     description = CASE WHEN id = $1 THEN COALESCE($3,  description) ELSE description END,
	//     ...
	// WHERE id IN ($1, $7, ...)

	nameCases := make([]string, 0, len(updates))
	descCases := make([]string, 0, len(updates))
	priceCases := make([]string, 0, len(updates))
	stockCases := make([]string, 0, len(updates))
	categoryCases := make([]string, 0, len(updates))
	inClause := make([]string, 0, len(updates))
	args := make([]any, 0, len(updates)*6)

	argIdx := 1
	for id, input := range updates {
		idPlaceholder := fmt.Sprintf("$%d", argIdx)

		nameCases = append(nameCases, fmt.Sprintf("WHEN id = %s THEN COALESCE($%d, name)", idPlaceholder, argIdx+1))
		descCases = append(descCases, fmt.Sprintf("WHEN id = %s THEN COALESCE($%d, description)", idPlaceholder, argIdx+2))
		priceCases = append(priceCases, fmt.Sprintf("WHEN id = %s THEN COALESCE($%d, price)", idPlaceholder, argIdx+3))
		stockCases = append(stockCases, fmt.Sprintf("WHEN id = %s THEN COALESCE($%d, stock)", idPlaceholder, argIdx+4))
		categoryCases = append(categoryCases, fmt.Sprintf("WHEN id = %s THEN COALESCE($%d, category_id)", idPlaceholder, argIdx+5))
		inClause = append(inClause, idPlaceholder)

		args = append(args, id, input.Name, input.Description, input.Price, input.Stock, input.CategoryID)
		argIdx += 6
	}

	query := fmt.Sprintf(`
		UPDATE products
		SET
			name        = CASE %s ELSE name        END,
			description = CASE %s ELSE description END,
			price       = CASE %s ELSE price       END,
			stock       = CASE %s ELSE stock       END,
			category_id = CASE %s ELSE category_id END,
			updated_at  = NOW()
		WHERE id IN (%s)
		RETURNING id, name, description, price, stock, category_id, created_at, updated_at
	`,
		strings.Join(nameCases, " "),
		strings.Join(descCases, " "),
		strings.Join(priceCases, " "),
		strings.Join(stockCases, " "),
		strings.Join(categoryCases, " "),
		strings.Join(inClause, ", "),
	)

	rows, err := r.db.QueryContext(ctx, query, args...)

	if err != nil {
		return fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []*entity.Product
	for rows.Next() {
		p := &entity.Product{}
		var desc sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&desc,
			&p.Price,
			&p.Stock,
			&p.CategoryID,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return fmt.Errorf("failed to scan product: %w", err)
		}

		if desc.Valid {
			p.Description = &desc.String
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows error %w", err)
	}

	return nil
}

func (r *PostgresProductRepository) Delete(ctx context.Context, id int64) error {
	query := `
		DELETE FROM products
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, id)

	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &apperrors.NotFoundError{Resource: "Product"}
	}

	return nil
}

func (r *PostgresProductRepository) GetByIDs(ctx context.Context, ids []int64) ([]entity.Product, error) {
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, name, description, price, stock, category_id, created_at, updated_at
		FROM products
		WHERE id IN (%s)
	`, strings.Join(placeholders, ", "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []entity.Product
	for rows.Next() {
		var p entity.Product
		var desc sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&desc,
			&p.Price,
			&p.Stock,
			&p.CategoryID,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		if desc.Valid {
			p.Description = &desc.String
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}
