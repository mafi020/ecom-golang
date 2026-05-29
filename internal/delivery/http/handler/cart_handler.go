package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/delivery/http/request"
	"github.com/mafi020/ecom-golang/internal/delivery/http/utils"
	"github.com/mafi020/ecom-golang/internal/entity"
	"github.com/mafi020/ecom-golang/internal/response"
)

type cartUsecase interface {
	GetCart(ctx context.Context, userID int64) (*entity.Cart, error)
	AddItem(ctx context.Context, userID, productID int64, quantity int) (*entity.Cart, error)
	UpdateItem(ctx context.Context, userID, productID int64, quantity int) (*entity.Cart, error)
	RemoveItem(ctx context.Context, userID, productID int64) (*entity.Cart, error)
	ClearCart(ctx context.Context, userID int64) error
}

type CartHandler struct {
	cartUsecase cartUsecase
}

func NewCartHandler(cartUsecase cartUsecase) *CartHandler {
	return &CartHandler{cartUsecase: cartUsecase}
}

func (h *CartHandler) GetCart(c *gin.Context) {
	userID := c.GetInt64("user_id")

	cart, err := h.cartUsecase.GetCart(c.Request.Context(), userID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) AddItem(c *gin.Context) {

	var req request.CartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, apperrors.ParseValidationError(err))
		return
	}

	userID := c.GetInt64("user_id")

	cart, err := h.cartUsecase.AddItem(c.Request.Context(), userID, req.ProductID, req.Quantity)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Success(c, cart)
}

func (h *CartHandler) UpdateItem(c *gin.Context) {
	productID, err := utils.ParseID(c, "product_id")
	if err != nil {
		utils.HandleError(c, &apperrors.BadRequestError{Errors: map[string]string{"product": "Invalid Product ID"}})
		return
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, apperrors.ParseValidationError(err))
		return
	}

	userID := c.GetInt64("user_id")

	cart, err := h.cartUsecase.UpdateItem(c.Request.Context(), userID, productID, req.Quantity)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Success(c, cart)
}

func (h *CartHandler) RemoveItem(c *gin.Context) {

	productID, err := utils.ParseID(c, "product_id")
	if err != nil {
		utils.HandleError(c, &apperrors.BadRequestError{Errors: map[string]string{"product": "Invalid Product ID"}})
		return
	}

	userID := c.GetInt64("user_id")
	cart, err := h.cartUsecase.RemoveItem(c.Request.Context(), userID, productID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Success(c, cart)
}

func (h *CartHandler) ClearCart(c *gin.Context) {
	userID := c.GetInt64("user_id")

	if err := h.cartUsecase.ClearCart(c.Request.Context(), userID); err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Message(c, "Cart Cleared")
}
