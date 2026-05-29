package request

type CartRequest struct {
	Items []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
}
