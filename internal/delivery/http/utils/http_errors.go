package utils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mafi020/ecom-golang/internal/apperrors"
)

func HandleError(c *gin.Context, err error) {
	var notFound *apperrors.NotFoundError
	var conflict *apperrors.ConflictError
	var validation *apperrors.ValidationError
	var badRequest *apperrors.BadRequestError
	var unauthorized *apperrors.UnauthorizedError
	var forbidden *apperrors.ForbiddenError
	var tooManyRequests *apperrors.TooManyRequestsError

	switch {
	case errors.As(err, &notFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.As(err, &conflict):
		c.JSON(http.StatusConflict, gin.H{"errors": conflict.Errors})
	case errors.As(err, &validation):
		c.JSON(http.StatusBadRequest, gin.H{"errors": validation.Errors})
	case errors.As(err, &badRequest):
		c.JSON(http.StatusBadRequest, gin.H{"errors": badRequest.Errors})
	case errors.As(err, &unauthorized):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.As(err, &forbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.As(err, &tooManyRequests):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}

// **** Usage Gudie on different Layers ****

// ── HANDLER — binding error ──────────────────────────────
// if err := c.ShouldBindJSON(&req); err != nil {
//     apperrors.HandleError(c, apperrors.ParseValidationError(err))
//     return
// }

// // ── HANDLER — usecase error ──────────────────────────────
// product, err := h.productUsecase.GetByID(c.Request.Context(), id)
// if err != nil {
//     apperrors.HandleError(c, err)
//     return
// }

// // ── USECASE — not found ──────────────────────────────────
// product, err := uc.productRepo.GetByID(ctx, id)
// if err != nil {
//     return nil, &apperrors.NotFoundError{Resource: "product", ID: id}
// }

// // ── USECASE — business validation ───────────────────────
// if product.Stock < item.Quantity {
//     return nil, &apperrors.ValidationError{Errors: map[string]string{
//         "stock": fmt.Sprintf("insufficient stock: have %d, want %d", product.Stock, item.Quantity),
//     }}
// }

// // ── USECASE — unauthorized ───────────────────────────────
// if order.UserID != userID {
//     return nil, &apperrors.UnauthorizedError{Message: "order not found"}
// }

// // ── USECASE — forbidden ──────────────────────────────────
// if user.Role != "admin" {
//     return nil, &apperrors.ForbiddenError{Message: "admin access required"}
// }

// // ── REPO — conflict ──────────────────────────────────────
// return apperrors.HandleUniqueViolation(err, map[string]string{
//     "users_email_key": "email",
//     "users_slug_key":  "slug",
// })
