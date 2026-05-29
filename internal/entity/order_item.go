package entity

import "time"

type OrderItem struct {
	ID        int64     `json:"id"`
	OrderID   int64     `json:"order_id"`
	ProductID int64     `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Product *Product `json:"product,omitempty"`
	Order   *Order   `json:"order,omitempty"`
}
