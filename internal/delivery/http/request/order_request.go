package request

type OrderRequest struct {
	Items []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
}
