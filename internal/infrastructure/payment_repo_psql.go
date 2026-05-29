package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/entity"
)

type PostgresPaymentRepository struct {
	db *sql.DB
}

func NewPostgresPaymentRepository(db *sql.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db: db}
}

// Create persists the core ledger entry using our universal internal transaction UUID
func (r *PostgresPaymentRepository) Create(ctx context.Context, payment *entity.Payment) (*entity.Payment, error) {
	query := `
		INSERT INTO payments (order_id, transaction_id, method, status, amount)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, order_id, transaction_id, method, status, amount, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		payment.OrderID,
		payment.TransactionID, // Universal identity reference mapping
		payment.Method,
		payment.Status,
		payment.Amount,
	).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.TransactionID,
		&payment.Method,
		&payment.Status,
		&payment.Amount,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		return nil, apperrors.HandleUniqueViolation(err, map[string]string{
			"payments_order_id_key": "order_id",
		})
	}

	return payment, nil
}

// CreateOnlineTransaction maps third party verification provider payloads
func (r *PostgresPaymentRepository) CreateOnlineTransaction(ctx context.Context, tx *entity.OnlineTransaction) (*entity.OnlineTransaction, error) {
	query := `
		INSERT INTO online_transactions (payment_id, provider, gateway_ref, gateway_status, raw_response)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, payment_id, provider, gateway_ref, gateway_status, raw_response, created_at
	`

	var (
		gatewayRef    sql.NullString
		gatewayStatus sql.NullString
		rawResponse   []byte
	)

	err := r.db.QueryRowContext(ctx, query,
		tx.PaymentID,
		tx.Provider,
		tx.GatewayRef,
		tx.GatewayStatus,
		tx.RawResponse,
	).Scan(
		&tx.ID,
		&tx.PaymentID,
		&tx.Provider,
		&gatewayRef,
		&gatewayStatus,
		&rawResponse,
		&tx.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create online transaction: %w", err)
	}

	if gatewayRef.Valid {
		tx.GatewayRef = gatewayRef.String
	}
	if gatewayStatus.Valid {
		tx.GatewayStatus = gatewayStatus.String
	}
	if rawResponse != nil {
		tx.RawResponse = json.RawMessage(rawResponse)
	}

	return tx, nil
}

// CreateCODDetail logs downstream physical checkout verification hooks
func (r *PostgresPaymentRepository) CreateCODDetail(ctx context.Context, detail *entity.CODDetail) (*entity.CODDetail, error) {
	query := `
		INSERT INTO cod_details (payment_id, collected_at)
		VALUES ($1, $2)
		RETURNING id, payment_id, collected_at, created_at, updated_at
	`

	var collectedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query,
		detail.PaymentID,
		detail.CollectedAt,
	).Scan(
		&detail.ID,
		&detail.PaymentID,
		&collectedAt,
		&detail.CreatedAt,
		&detail.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create cod detail: %w", err)
	}

	if collectedAt.Valid {
		detail.CollectedAt = &collectedAt.Time
	}

	return detail, nil
}

// GetPaymentByOrderID fetches payment metadata using a single row scan
func (r *PostgresPaymentRepository) GetPaymentByOrderID(ctx context.Context, orderID int64) (*entity.Payment, error) {
	query := `
		SELECT p.id, p.order_id, p.transaction_id, p.method, p.status, p.amount, p.created_at, p.updated_at,
			   ot.id, ot.provider, ot.gateway_ref, ot.gateway_status, ot.raw_response, ot.created_at,
			   cd.id, cd.collected_at, cd.created_at, cd.updated_at
		FROM payments p
		LEFT JOIN online_transactions ot ON ot.payment_id = p.id AND p.method = 'online'
		LEFT JOIN cod_details cd         ON cd.payment_id = p.id AND p.method = 'cod'
		WHERE p.order_id = $1
	`

	var (
		// core variables
		paymentID     int64
		pOrderID      int64
		txUUID        string
		paymentMethod string
		paymentStatus string
		paymentAmount float64
		paymentCA     sql.NullTime
		paymentUA     sql.NullTime

		// online relations
		otID            sql.NullInt64
		otProvider      sql.NullString
		otGatewayRef    sql.NullString
		otGatewayStatus sql.NullString
		otRawResponse   []byte
		otCA            sql.NullTime

		// cod relations
		cdID          sql.NullInt64
		cdCollectedAt sql.NullTime
		cdCA          sql.NullTime
		cdUA          sql.NullTime
	)

	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&paymentID, &pOrderID, &txUUID, &paymentMethod, &paymentStatus, &paymentAmount, &paymentCA, &paymentUA,
		&otID, &otProvider, &otGatewayRef, &otGatewayStatus, &otRawResponse, &otCA,
		&cdID, &cdCollectedAt, &cdCA, &cdUA,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &apperrors.NotFoundError{Resource: "payment"}
		}
		return nil, fmt.Errorf("failed to scan payment: %w", err)
	}

	payment := &entity.Payment{
		ID:            paymentID,
		OrderID:       pOrderID,
		TransactionID: txUUID,
		Method:        entity.PaymentMethod(paymentMethod),
		Status:        entity.PaymentStatus(paymentStatus),
		Amount:        paymentAmount,
		CreatedAt:     paymentCA.Time,
		UpdatedAt:     paymentUA.Time,
	}

	if otID.Valid {
		ot := &entity.OnlineTransaction{
			ID:        otID.Int64,
			PaymentID: paymentID,
			CreatedAt: otCA.Time,
		}
		if otProvider.Valid {
			ot.Provider = entity.PaymentProvider(otProvider.String)
		}
		if otGatewayRef.Valid {
			ot.GatewayRef = otGatewayRef.String
		}
		if otGatewayStatus.Valid {
			ot.GatewayStatus = otGatewayStatus.String
		}
		if otRawResponse != nil {
			ot.RawResponse = json.RawMessage(otRawResponse)
		}
		payment.OnlineTransaction = ot
	}

	if cdID.Valid {
		cd := &entity.CODDetail{
			ID:        cdID.Int64,
			PaymentID: paymentID,
			CreatedAt: cdCA.Time,
			UpdatedAt: cdUA.Time,
		}
		if cdCollectedAt.Valid {
			t := cdCollectedAt.Time
			cd.CollectedAt = &t
		}
		payment.CODDetail = cd
	}

	return payment, nil
}

// UpdateStatus changes the state of the payment transaction ledger row
func (r *PostgresPaymentRepository) UpdateStatus(ctx context.Context, id int64, status entity.PaymentStatus) error {
	query := `UPDATE payments SET status = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return &apperrors.NotFoundError{Resource: "payment"}
	}

	return nil
}

// UpdateCODDetail validates and closes financial accounting milestones
func (r *PostgresPaymentRepository) UpdateCODDetail(ctx context.Context, paymentID int64, detail *entity.CODDetail) (*entity.CODDetail, error) {
	query := `
		UPDATE cod_details
		SET collected_at = COALESCE($1, collected_at),
		    updated_at   = NOW()
		WHERE payment_id = $2
		RETURNING id, payment_id, collected_at, created_at, updated_at
	`

	var collectedAt sql.NullTime

	// Completed the missing scan mapping parameters cleanly
	err := r.db.QueryRowContext(ctx, query, detail.CollectedAt, paymentID).Scan(
		&detail.ID,
		&detail.PaymentID,
		&collectedAt,
		&detail.CreatedAt,
		&detail.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &apperrors.NotFoundError{Resource: "cod_detail"}
		}
		return nil, fmt.Errorf("failed to update cod detail: %w", err)
	}

	if collectedAt.Valid {
		detail.CollectedAt = &collectedAt.Time
	}

	return detail, nil
}
