package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mafi020/ecom-golang/internal/entity"
)

type PostgresOrderItemRepository struct {
	db *sql.DB
}

func NewPostgresOrderItemRepository(db *sql.DB) *PostgresOrderItemRepository {
	return &PostgresOrderItemRepository{db: db}
}

func (r *PostgresOrderItemRepository) CreateOrderItems(ctx context.Context, orderID int64, items []entity.OrderItem) ([]entity.OrderItem, error) {
	db := GetExecutor(ctx, r.db)

	query := `
        INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())
        RETURNING id, order_id, product_id, quantity, price, created_at, updated_at
    `

	for i := range items {
		err := db.QueryRowContext(
			ctx,
			query,
			orderID,
			items[i].ProductID,
			items[i].Quantity,
			items[i].Price,
		).Scan(
			&items[i].ID,
			&items[i].OrderID,
			&items[i].ProductID,
			&items[i].Quantity,
			&items[i].Price,
			&items[i].CreatedAt,
			&items[i].UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	return items, nil
}

func (r *PostgresOrderItemRepository) GetOrderItemsByOrderID(ctx context.Context, orderID int64) ([]entity.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, quantity, price, created_at, updated_at
		FROM order_items
		WHERE order_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()
	var items []entity.OrderItem
	for rows.Next() {
		var item entity.OrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Price,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}
