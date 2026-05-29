package entity

import (
	"encoding/json"
	"time"
)

type PaymentMethod string

const (
	PaymentMethodCOD    PaymentMethod = "cod"
	PaymentMethodOnline PaymentMethod = "online"
)

type PaymentProvider string

const (
	PaymentProviderStripe     PaymentProvider = "stripe"
	PaymentProviderSSLCOMMERZ PaymentProvider = "sslcommerz"
	PaymentProviderPaypal     PaymentProvider = "paypal"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// ── 1. THE MAIN LEDGER ENTRY ──────────────────────────────────────────────────
type Payment struct {
	ID      int64 `json:"id"`
	OrderID int64 `json:"order_id"`
	// PROVEN FIX: Internal universal transaction UUID moved here to the parent root.
	// It is generated at checkout instantly for both COD and Online.
	TransactionID string        `json:"transaction_id"`
	Method        PaymentMethod `json:"method"`
	Status        PaymentStatus `json:"status"`
	Amount        float64       `json:"amount"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`

	// Sub-relations containing ONLY type-specific meta payloads
	OnlineTransaction *OnlineTransaction `json:"online_transaction,omitempty"`
	CODDetail         *CODDetail         `json:"cod_detail,omitempty"`
}

// ── 2. ONLINE METADATA ONLY ───────────────────────────────────────────────────
type OnlineTransaction struct {
	ID            int64           `json:"id"`
	PaymentID     int64           `json:"payment_id"`
	Provider      PaymentProvider `json:"provider"`
	GatewayRef    string          `json:"gateway_ref"`    // Gateway's checkout session ID
	GatewayStatus string          `json:"gateway_status"` // e.g., "succeeded"
	RawResponse   json.RawMessage `json:"raw_response"`   // Full gateway webhook payload
	CreatedAt     time.Time       `json:"created_at"`
}

// ── 3. COD FINANCIAL LEDGER ONLY ──────────────────────────────────────────────
type CODDetail struct {
	ID          int64      `json:"id"`
	PaymentID   int64      `json:"payment_id"`
	CollectedAt *time.Time `json:"collected_at,omitempty"` // The moment the cash is secure
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
