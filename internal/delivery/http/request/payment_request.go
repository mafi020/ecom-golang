package request

import (
	"encoding/json"

	"github.com/mafi020/ecom-golang/internal/entity"
)

type CreatePaymentRequest struct {
	OrderID       int64  `json:"order_id"       binding:"required,gt=0"`
	Provider      string `json:"provider"       binding:"required,oneof=manual stripe sslcommerz paypal"`
	TransactionID string `json:"transaction_id"`
}

type PayOnlineRequest struct {
	OrderID       int64                  `json:"order_id"       binding:"required,gt=0"`
	Provider      entity.PaymentProvider `json:"provider"       binding:"required,oneof=stripe paypal sslcommerz"`
	GatewayRef    string                 `json:"gateway_ref"`
	GatewayStatus string                 `json:"gateway_status"`
	RawResponse   json.RawMessage        `json:"raw_response"`
}
