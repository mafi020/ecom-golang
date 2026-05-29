package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/delivery/http/request"
	"github.com/mafi020/ecom-golang/internal/delivery/http/utils"
	"github.com/mafi020/ecom-golang/internal/entity"
	"github.com/mafi020/ecom-golang/internal/response"
)

type paymentUsecase interface {
	PayOnline(ctx context.Context, userID int64, orderID int64, provider entity.PaymentProvider, gatewayRef string, gatewayStatus string, rawResponse []byte) (*entity.Payment, error)
	PayCOD(ctx context.Context, userID int64, orderID int64) (*entity.Payment, error)
	CollectCOD(ctx context.Context, orderID int64) (*entity.Payment, error)
	GetPaymentByOrderID(ctx context.Context, orderID, userID int64) (*entity.Payment, error)
}

type PaymentHandler struct {
	paymentUsecase paymentUsecase
}

func NewPaymentHandler(uc paymentUsecase) *PaymentHandler {
	return &PaymentHandler{paymentUsecase: uc}
}

func (h *PaymentHandler) PayOnline(c *gin.Context) {
	userID := c.GetInt64("userID")

	var req request.PayOnlineRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, apperrors.ParseValidationError(err))
		return
	}

	payment, err := h.paymentUsecase.PayOnline(
		c.Request.Context(),
		userID,
		req.OrderID,
		entity.PaymentProvider(req.Provider),
		req.GatewayRef,
		req.GatewayStatus,
		req.RawResponse,
	)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Created(c, payment)
}

func (h *PaymentHandler) PayCOD(c *gin.Context) {
	userID := c.GetInt64("userID")

	var req struct {
		OrderID        int64   `json:"order_id"        binding:"required,gt=0"`
		CourierPartner *string `json:"courier_partner"`
		TrackingNumber *string `json:"tracking_number"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, apperrors.ParseValidationError(err))
		return
	}

	payment, err := h.paymentUsecase.PayCOD(c.Request.Context(), userID, req.OrderID)

	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Created(c, payment)
}

// Admin only
func (h *PaymentHandler) CollectCOD(c *gin.Context) {
	orderID, err := utils.ParseID(c, "order_id")
	if err != nil {
		utils.HandleError(c, &apperrors.BadRequestError{Errors: map[string]string{"order": "Invalid Order ID"}})
		return
	}

	payment, err := h.paymentUsecase.CollectCOD(c.Request.Context(), orderID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Success(c, payment)
}

func (h *PaymentHandler) GetPaymentByOrderID(c *gin.Context) {
	userID := c.GetInt64("userID")

	orderID, err := utils.ParseID(c, "order_id")
	if err != nil {
		utils.HandleError(c, &apperrors.BadRequestError{Errors: map[string]string{"order": "Invalid Order ID"}})
		return
	}

	payment, err := h.paymentUsecase.GetPaymentByOrderID(c.Request.Context(), orderID, userID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Success(c, payment)
}
