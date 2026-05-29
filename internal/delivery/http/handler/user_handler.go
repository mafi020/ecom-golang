package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/delivery/http/utils"
	"github.com/mafi020/ecom-golang/internal/entity"
	"github.com/mafi020/ecom-golang/internal/response"
)

type userUseCase interface {
	GetUserByID(ctx context.Context, id int64) (*entity.User, error)
	GetUsers(ctx context.Context, params entity.GetUsersParams) ([]entity.User, int, error)
	DeleteUser(ctx context.Context, id int64) error
}

type UserHandler struct {
	userUsecase userUseCase
}

func NewUserHandler(uc userUseCase) *UserHandler {
	return &UserHandler{userUsecase: uc}
}

// GET /users/:id
func (h *UserHandler) GetUserByID(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		utils.HandleError(c, &apperrors.BadRequestError{Errors: map[string]string{"user": "Invalid User ID"}})
		return
	}

	user, err := h.userUsecase.GetUserByID(c.Request.Context(), id)

	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) GetUsers(c *gin.Context) {

	role := c.Query("role")
	if role != "" && role != string(entity.RoleAdmin) && role != string(entity.RoleCustomer) {
		role = ""
	}

	req := utils.ParseQueryParams(c)

	params := entity.GetUsersParams{
		QueryParams: req,
		Role:        role,
	}

	users, total, err := h.userUsecase.GetUsers(c.Request.Context(), params)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	if len(users) == 0 {
		users = []entity.User{} // Return empty array instead of null
	}

	response.Paginated(c, users, total, params.Page, params.Limit)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {

	id, err := utils.ParseID(c, "id")
	if err != nil {
		utils.HandleError(c, &apperrors.BadRequestError{Errors: map[string]string{"user": "Invalid User ID"}})
		return
	}

	userID, _ := c.Get("user_id")
	if userID == id {
		utils.HandleError(c, &apperrors.ForbiddenError{Message: "Cannot delete your own account"})
		return
	}

	err = h.userUsecase.DeleteUser(c.Request.Context(), id)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Message(c, "User deleted")
}
