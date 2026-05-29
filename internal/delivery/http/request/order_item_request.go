package request

type OrderItemRequest struct {
	ProductID int64   `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,gte=1"`
	Price     float64 `json:"price" binding:"required,gte=1.0"`
}
