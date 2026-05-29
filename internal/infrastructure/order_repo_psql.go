package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/entity"
)

type PostgresOrderRepository struct {
	db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
	return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) CreateOrder(ctx context.Context, order *entity.Order) error {
	db := GetExecutor(ctx, r.db)

	query := `
		INSERT INTO orders (user_id, status, total_price, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, user_id, status, total_price, created_at, updated_at 
	`

	var orderID int64

	err := db.QueryRowContext(
		ctx,
		query,
		order.UserID,
		order.Status,
		order.TotalPrice,
	).Scan(&orderID,
		&order.UserID,
		&order.Status,
		&order.TotalPrice,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	order.ID = orderID

	return nil
}

func (r *PostgresOrderRepository) ListOrders(ctx context.Context, userID *int64, params entity.GetOrdersParams) ([]entity.Order, int, error) {
	// 1. Base query
	baseQuery := `FROM orders WHERE 1=1`
	args := []any{}
	argIndex := 1

	if userID != nil {
		baseQuery += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *userID)
		argIndex++
	}
	// 4. Get total count
	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery

	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to fetch orders %w", err)
	}
	allowedSortColumns := map[string]bool{
		"created_at": true,
	}
	if !allowedSortColumns[params.SortBy] {
		params.SortBy = "created_at"
	}
	if params.SortOrder != "asc" {
		params.SortOrder = "desc"
	}
	mainQuery := fmt.Sprintf(
		"SELECT id, user_id, status, total_price, created_at, updated_at %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		baseQuery, params.SortBy, params.SortOrder, argIndex, argIndex+1,
	)
	args = append(args, params.Limit, params.Offset())
	rows, err := r.db.QueryContext(ctx, mainQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query orders %w", err)
	}
	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var o entity.Order
		err := rows.Scan(
			&o.ID, &o.UserID, &o.Status, &o.TotalPrice,
			&o.CreatedAt, &o.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("rows error %w", err)
		}

		orders = append(orders, o)
	}
	return orders, total, nil
}

func (r *PostgresOrderRepository) GetOrderByID(ctx context.Context, id, userID int64) (*entity.Order, error) {
	query := `
		SELECT o.id, o.user_id, o.status, o.total_price, o.created_at, o.updated_at,
			   oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.created_at, oi.updated_at
		FROM orders o
		LEFT JOIN order_items oi ON oi.order_id = o.id
		WHERE o.id = $1 AND o.user_id = $2
	`

	rows, err := r.db.QueryContext(ctx, query, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order: %w", err)
	}
	defer rows.Close()

	var order *entity.Order
	seenItems := make(map[int64]bool) // Tracks the duplicate items

	for rows.Next() {
		var (
			// order item variables
			itemID    sql.NullInt64
			orderID   sql.NullInt64
			productID sql.NullInt64
			quantity  sql.NullInt32
			itemPrice sql.NullFloat64
			itemCA    sql.NullTime
			itemUA    sql.NullTime
		)

		if order == nil {
			order = &entity.Order{OrderItems: []entity.OrderItem{}}
		}

		err := rows.Scan(
			// Order core data fields
			&order.ID, &order.UserID, &order.Status, &order.TotalPrice, &order.CreatedAt, &order.UpdatedAt,
			// Order Item nullable columns
			&itemID, &orderID, &productID, &quantity, &itemPrice, &itemCA, &itemUA,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Map items cleanly avoiding query duplicates
		if itemID.Valid && !seenItems[itemID.Int64] {
			seenItems[itemID.Int64] = true
			order.OrderItems = append(order.OrderItems, entity.OrderItem{
				ID:        itemID.Int64,
				OrderID:   orderID.Int64,
				ProductID: productID.Int64,
				Quantity:  int(quantity.Int32),
				Price:     itemPrice.Float64,
				CreatedAt: itemCA.Time,
				UpdatedAt: itemUA.Time,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	if order == nil {
		return nil, &apperrors.NotFoundError{Resource: "order"}
	}

	return order, nil
}

func (r *PostgresOrderRepository) GetOrdersByUserID(ctx context.Context, userID int64, params entity.GetOrdersParams) ([]entity.Order, int, error) {
	baseQuery := `FROM orders WHERE user_id = $1`
	args := []any{userID}
	argIndex := 2

	// Filter by status
	if params.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, params.Status)
		argIndex++
	}

	// Total count
	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) "+baseQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// Sort safety
	allowedSortColumns := map[string]bool{
		"created_at":  true,
		"total_price": true,
		"status":      true,
	}
	if !allowedSortColumns[params.SortBy] {
		params.SortBy = "created_at"
	}
	if params.SortOrder != "asc" {
		params.SortOrder = "desc"
	}

	mainQuery := fmt.Sprintf(
		"SELECT id, user_id, status, total_price, created_at, updated_at %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		baseQuery, params.SortBy, params.SortOrder, argIndex, argIndex+1,
	)
	args = append(args, params.Limit, params.Offset())

	rows, err := r.db.QueryContext(ctx, mainQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var o entity.Order
		if err := rows.Scan(
			&o.ID, &o.UserID, &o.Status, &o.TotalPrice, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return orders, total, nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id int64, status entity.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return &apperrors.NotFoundError{Resource: "order"}
	}

	return nil
}
