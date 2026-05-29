package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/entity"
)

type paymentRepo interface {
	Create(ctx context.Context, payment *entity.Payment) (*entity.Payment, error)
	CreateOnlineTransaction(ctx context.Context, tx *entity.OnlineTransaction) (*entity.OnlineTransaction, error)
	CreateCODDetail(ctx context.Context, detail *entity.CODDetail) (*entity.CODDetail, error)
	GetPaymentByOrderID(ctx context.Context, orderID int64) (*entity.Payment, error)
	UpdateStatus(ctx context.Context, id int64, status entity.PaymentStatus) error
	UpdateCODDetail(ctx context.Context, paymentID int64, detail *entity.CODDetail) (*entity.CODDetail, error)
}

type paymentOrderRepo interface {
	GetOrderByID(ctx context.Context, id, userID int64) (*entity.Order, error)
	UpdateStatus(ctx context.Context, id int64, status entity.OrderStatus) error
}

type tm interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}

type PaymentUseCase struct {
	paymentRepo paymentRepo
	orderRepo   paymentOrderRepo
	tm          tm
}

func NewPaymentUseCase(paymentRepo paymentRepo, orderRepo paymentOrderRepo, tm tm) *PaymentUseCase {
	return &PaymentUseCase{paymentRepo: paymentRepo, orderRepo: orderRepo, tm: tm}
}

// ── Online Payment ────────────────────────────────────────────────────────────

type OnlinePaymentInput struct {
	OrderID       int64
	Provider      entity.PaymentProvider
	GatewayRef    string
	GatewayStatus string
	RawResponse   json.RawMessage
}

func (uc *PaymentUseCase) PayOnline(
	ctx context.Context,
	userID int64,
	orderID int64,
	provider entity.PaymentProvider,
	gatewayRef string,
	gatewayStatus string,
	rawResponse []byte,
) (*entity.Payment, error) {
	order, err := uc.orderRepo.GetOrderByID(ctx, orderID, userID)
	if err != nil {
		return nil, err
	}

	if order.Status != entity.OrderStatusPending {
		return nil, &apperrors.ValidationError{Errors: map[string]string{
			"status": fmt.Sprintf("order must be pending for online payment, current: %s", order.Status),
		}}
	}

	var payment *entity.Payment

	err = uc.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		// 1. Create payment record (Generate universal internal tracking ID directly on root)
		txUUID := uuid.New().String()
		payment, err = uc.paymentRepo.Create(txCtx, &entity.Payment{
			OrderID:       order.ID,
			TransactionID: txUUID,
			Method:        entity.PaymentMethodOnline,
			Status:        entity.PaymentStatusCompleted, // Direct instantiation as completed
			Amount:        order.TotalPrice,
		})
		if err != nil {
			return err
		}

		// 2. Create online gateway specific tracking transaction info
		onlineTx, err := uc.paymentRepo.CreateOnlineTransaction(txCtx, &entity.OnlineTransaction{
			PaymentID:     payment.ID,
			Provider:      provider,
			GatewayRef:    gatewayRef,
			GatewayStatus: gatewayStatus,
			RawResponse:   rawResponse,
		})
		if err != nil {
			return err
		}
		payment.OnlineTransaction = onlineTx

		// 3. Update parent order fulfillment status
		return uc.orderRepo.UpdateStatus(txCtx, order.ID, entity.OrderStatusPaid)
	})

	return payment, err
}

// ── COD Payment ───────────────────────────────────────────────────────────────

func (uc *PaymentUseCase) PayCOD(ctx context.Context, userID, orderID int64) (*entity.Payment, error) {
	order, err := uc.orderRepo.GetOrderByID(ctx, orderID, userID)
	if err != nil {
		return nil, err
	}

	if order.Status != entity.OrderStatusPending {
		return nil, &apperrors.ValidationError{Errors: map[string]string{
			"status": fmt.Sprintf("order must be pending for COD, current: %s", order.Status),
		}}
	}

	var payment *entity.Payment

	err = uc.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		// 1. Create payment ledger entry (Stays pending until delivery collection)
		txUUID := uuid.New().String()
		payment, err = uc.paymentRepo.Create(txCtx, &entity.Payment{
			OrderID:       order.ID,
			TransactionID: txUUID,
			Method:        entity.PaymentMethodCOD,
			Status:        entity.PaymentStatusPending,
			Amount:        order.TotalPrice,
		})
		if err != nil {
			return err
		}

		// 2. Create matching empty cash-ledger detail row
		codDetail, err := uc.paymentRepo.CreateCODDetail(txCtx, &entity.CODDetail{
			PaymentID: payment.ID,
		})
		if err != nil {
			return err
		}
		payment.CODDetail = codDetail

		return nil
	})

	return payment, err
}

// ── COD Collection (admin) ────────────────────────────────────────────────────

func (uc *PaymentUseCase) CollectCOD(ctx context.Context, orderID int64) (*entity.Payment, error) {
	payment, err := uc.paymentRepo.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if payment.Method != entity.PaymentMethodCOD {
		return nil, &apperrors.ValidationError{Errors: map[string]string{
			"method": "payment is not COD",
		}}
	}

	if payment.Status == entity.PaymentStatusCompleted {
		return nil, &apperrors.ValidationError{Errors: map[string]string{
			"status": "payment already completed",
		}}
	}

	now := time.Now()

	err = uc.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		// 1. Mark financial ledger row as reconciled (money in hand)
		codDetail, err := uc.paymentRepo.UpdateCODDetail(txCtx, payment.ID, &entity.CODDetail{
			CollectedAt: &now,
		})
		if err != nil {
			return err
		}
		payment.CODDetail = codDetail

		// 2. Complete payment status
		if err := uc.paymentRepo.UpdateStatus(txCtx, payment.ID, entity.PaymentStatusCompleted); err != nil {
			return err
		}
		payment.Status = entity.PaymentStatusCompleted

		// 3. Chronologically progress order state machine (shipped -> delivered)
		return uc.orderRepo.UpdateStatus(txCtx, orderID, entity.OrderStatusDelivered)
	})

	return payment, err
}

func (uc *PaymentUseCase) GetPaymentByOrderID(ctx context.Context, orderID, userID int64) (*entity.Payment, error) {
	_, err := uc.orderRepo.GetOrderByID(ctx, orderID, userID)
	if err != nil {
		return nil, err
	}

	return uc.paymentRepo.GetPaymentByOrderID(ctx, orderID)
}
