package entity

import (
	"slices"
	"time"

	"github.com/mafi020/ecom-golang/internal/utils"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Updated matrix to enforce strictly forward chronological transitions
var validTransitions = map[OrderStatus][]OrderStatus{
	OrderStatusPending:   {OrderStatusPaid, OrderStatusShipped, OrderStatusCancelled},
	OrderStatusPaid:      {OrderStatusShipped, OrderStatusCancelled},
	OrderStatusShipped:   {OrderStatusDelivered, OrderStatusCancelled}, // FIXED: Removed backward transitions
	OrderStatusDelivered: {OrderStatusCompleted},                       // FIXED: Removed backward transitions
	OrderStatusCompleted: {},
	OrderStatusCancelled: {},
}

func (s OrderStatus) CanTransitionTo(next OrderStatus) bool {
	allowed, ok := validTransitions[s]
	if !ok {
		return false
	}
	return slices.Contains(allowed, next)
}

type Order struct {
	ID         int64       `json:"id"`
	UserID     int64       `json:"user_id"`
	Status     OrderStatus `json:"status" default:"pending"`
	TotalPrice float64     `json:"total_price"`

	// PROVEN FIX: Logistics details live safely here inside the Order domain
	CourierPartner *string    `json:"courier_partner,omitempty"`
	TrackingNumber *string    `json:"tracking_number,omitempty"`
	ShippedAt      *time.Time `json:"shipped_at,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	OrderItems []OrderItem `json:"order_items,omitempty"`
	User       *User       `json:"user,omitempty"`
	Payment    *Payment    `json:"payment,omitempty"`
}

type GetOrdersParams struct {
	utils.QueryParams
	Status string `form:"status"`
}
