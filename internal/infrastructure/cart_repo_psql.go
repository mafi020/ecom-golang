package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/entity"
)

type PostgresCartRepository struct {
	db *sql.DB
}

func NewPostgresCartRepository(db *sql.DB) *PostgresCartRepository {
	return &PostgresCartRepository{db: db}
}

func (r *PostgresCartRepository) GetOrCreateCart(ctx context.Context, userID int64) (*entity.Cart, error) {
	query := `
		INSERT INTO carts (user_id) 
		VALUES ($1) 
		ON CONFLICT (user_id) DO UPDATE SET updated_at = NOW() 
		RETURNING id, user_id, created_at, updated_at
	`
	var cart entity.Cart
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&cart.ID, &cart.UserID, &cart.CreatedAt, &cart.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "cart"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get or create cart: %w", err)
	}
	return &cart, nil
}

func (r *PostgresCartRepository) GetCartByUserID(ctx context.Context, userID int64) (*entity.Cart, error) {
	query := `
		SELECT c.id, c.user_id, c.created_at, c.updated_at,
			   ci.id, ci.cart_id, ci.product_id, ci.quantity, ci.price
		FROM carts c
		LEFT JOIN cart_items ci ON ci.cart_id = c.id
		WHERE c.user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cart: %w", err)
	}
	defer rows.Close()

	var cart *entity.Cart

	for rows.Next() {
		var (
			itemID    sql.NullInt64
			cartID    sql.NullInt64
			productID sql.NullInt64
			quantity  sql.NullInt32
			price     sql.NullFloat64
		)

		if cart == nil {
			cart = &entity.Cart{Items: []entity.CartItem{}}
		}

		err := rows.Scan(
			&cart.ID, &cart.UserID, &cart.CreatedAt, &cart.UpdatedAt,
			&itemID, &cartID, &productID, &quantity, &price,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cart: %w", err)
		}

		if itemID.Valid {
			cart.Items = append(cart.Items, entity.CartItem{
				ID:        itemID.Int64,
				CartID:    cartID.Int64,
				ProductID: productID.Int64,
				Quantity:  int(quantity.Int32),
				Price:     price.Float64,
			})
		}
	}

	if cart == nil {
		return nil, &apperrors.NotFoundError{Resource: "cart"}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return cart, nil
}
