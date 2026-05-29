package entity

import "time"

type Cart struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Items     []CartItem `json:"items,omitempty"`
}

// GetSubtotal calculates the combined monetary value of all items in the cart
func (c *Cart) GetSubtotal() float64 {
	var total float64
	for _, item := range c.Items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}

// GetTotalItemCount calculates the sum of all individual product units
func (c *Cart) GetTotalItemCount() int {
	var count int
	for _, item := range c.Items {
		count += item.Quantity
	}
	return count
}
