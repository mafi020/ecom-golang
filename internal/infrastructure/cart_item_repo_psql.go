package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/entity"
)

type PostgresCartItemRepository struct {
	db *sql.DB
}

func NewPostgresCartItemRepository(db *sql.DB) *PostgresCartItemRepository {
	return &PostgresCartItemRepository{db: db}
}

func (r *PostgresCartItemRepository) AddItem(ctx context.Context, cartID int64, item *entity.CartItem) (*entity.CartItem, error) {
	query := `
		INSERT INTO cart_items (cart_id, product_id, quantity, price) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (cart_id, product_id) 
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity, price = EXCLUDED.price
		RETURNING id, cart_id, product_id, quantity, price
	`
	var result entity.CartItem
	err := r.db.QueryRowContext(ctx, query, cartID, item.ProductID, item.Quantity, item.Price).Scan(
		&result.ID, &result.CartID, &result.ProductID, &result.Quantity, &result.Price,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add cart item: %w", err)
	}
	return &result, nil
}

func (r *PostgresCartItemRepository) UpdateItemQuantity(ctx context.Context, cartID, productID int64, quantity int) (*entity.CartItem, error) {
	// If the user wants to reduce item quantity to 0, they must call RemoveItem.
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero; use RemoveItem to delete")
	}

	query := `
		UPDATE cart_items 
		SET quantity = $1 
		WHERE cart_id = $2 AND product_id = $3
		RETURNING id, cart_id, product_id, quantity, price
	`
	var item entity.CartItem
	err := r.db.QueryRowContext(ctx, query, quantity, cartID, productID).Scan(
		&item.ID, &item.CartID, &item.ProductID, &item.Quantity, &item.Price,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &apperrors.NotFoundError{Resource: "cart item"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update cart item: %w", err)
	}
	return &item, nil
}

func (r *PostgresCartItemRepository) RemoveItem(ctx context.Context, cartID, productID int64) error {
	query := `DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`
	result, err := r.db.ExecContext(ctx, query, cartID, productID)
	if err != nil {
		return fmt.Errorf("failed to remove cart item: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return &apperrors.NotFoundError{Resource: "cart item"}
	}
	return nil
}

func (r *PostgresCartItemRepository) ClearCart(ctx context.Context, cartID int64) error {
	query := `DELETE FROM cart_items WHERE cart_id = $1`
	result, err := r.db.ExecContext(ctx, query, cartID)

	if err != nil {
		return fmt.Errorf("failed to clear cart items: %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &apperrors.NotFoundError{Resource: "Cart Items"}
	}

	return nil
}
