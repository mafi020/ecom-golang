package entity

type CartItem struct {
	ID        int64   `json:"id"`
	CartID    int64   `json:"cart_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}
