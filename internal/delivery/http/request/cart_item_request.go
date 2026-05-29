package request

type CartItemRequest struct {
	ProductID int64 `json:"product_id" binding:"required,gte=1"`
	Quantity  int   `json:"quantity" binding:"required,gte=1"`
}
