package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/delivery/http/utils"
	"github.com/mafi020/ecom-golang/internal/entity"
	"github.com/mafi020/ecom-golang/internal/response"
	"github.com/mafi020/ecom-golang/internal/usecase"
)

type orderHandler struct {
	orderUsecase *usecase.OrderUseCase
}

func NewOrderHandler(orderUsecase *usecase.OrderUseCase) *orderHandler {
	return &orderHandler{orderUsecase: orderUsecase}
}

func (h *orderHandler) PlaceOrder(c *gin.Context) {
	userID := c.GetInt64("user_id")

	order, err := h.orderUsecase.Checkout(c.Request.Context(), userID)

	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Success(c, order)
}

func (h *orderHandler) GetOrderByID(c *gin.Context) {
	userID := c.GetInt64("user_id")

	id, err := utils.ParseID(c, "id")
	if err != nil {
		utils.HandleError(c, &apperrors.BadRequestError{Errors: map[string]string{"order": "Invalid Order ID"}})
		return
	}

	order, err := h.orderUsecase.GetOrderByID(c.Request.Context(), id, userID)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Success(c, order)
}

func (h *orderHandler) GetOrdersByUserID(c *gin.Context) {
	userID := c.GetInt64("user_id")

	params := entity.GetOrdersParams{
		QueryParams: utils.ParseQueryParams(c),
		Status:      c.Query("status"),
	}

	orders, total, err := h.orderUsecase.GetOrdersByUserID(c.Request.Context(), userID, params)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Paginated(c, orders, total, params.Page, params.Limit)
}
