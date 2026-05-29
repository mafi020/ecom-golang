package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/delivery/http/request"
	"github.com/mafi020/ecom-golang/internal/delivery/http/utils"
	"github.com/mafi020/ecom-golang/internal/entity"
	"github.com/mafi020/ecom-golang/internal/response"
	"github.com/mafi020/ecom-golang/internal/usecase"
)

type CategoryHandler struct {
	categoryUseCase *usecase.CategoryUseCase
}

func NewCategoryHandler(categoryUseCase *usecase.CategoryUseCase) *CategoryHandler {
	return &CategoryHandler{categoryUseCase: categoryUseCase}
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var body request.CategoryRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.HandleError(c, apperrors.ParseValidationError(err))
		return
	}

	category, err := h.categoryUseCase.CreateCategory(c.Request.Context(), body.Name, body.ParentID)

	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Success(c, category)
}

func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {

	id, err := utils.ParseID(c, "id")
	if err != nil {
		utils.HandleError(c, &apperrors.BadRequestError{Errors: map[string]string{"category": "Invalid Category ID"}})
		return
	}

	category, err := h.categoryUseCase.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Success(c, category)
}

func (h *CategoryHandler) GetCategoryByIDWithProducts(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		utils.HandleError(c, &apperrors.BadRequestError{Errors: map[string]string{"category": "Invalid Category ID"}})
		return
	}

	category, err := h.categoryUseCase.GetCategoryByIDWithProducts(c.Request.Context(), id)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Success(c, category)
}

func (h *CategoryHandler) GetAllCategories(c *gin.Context) {
	parentID, err := utils.ParseID(c, "parent_id")
	if err != nil {
		utils.HandleError(c, &apperrors.BadRequestError{Errors: map[string]string{"category": "Invalid Category ID"}})
		return
	}

	params := entity.GetCategoriesParams{
		QueryParams: utils.ParseQueryParams(c),
		ParentID:    parentID,
	}

	categories, total, err := h.categoryUseCase.GetAllCategories(c.Request.Context(), params)

	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Paginated(c, categories, total, params.Page, params.Limit)
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {

	id, err := utils.ParseID(c, "id")
	if err != nil {
		utils.HandleError(c, &apperrors.BadRequestError{Errors: map[string]string{"category": "Invalid Category ID"}})
		return
	}

	var body request.CategoryRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.HandleError(c, apperrors.ParseValidationError(err))
		return
	}

	category := &entity.Category{
		ID:       id,
		Name:     body.Name,
		Slug:     utils.GenerateSlug(body.Name),
		ParentID: body.ParentID,
	}

	if err := h.categoryUseCase.UpdateCategory(c.Request.Context(), category); err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Success(c, category)
}

func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id, err := utils.ParseID(c, "id")
	if err != nil {
		utils.HandleError(c, &apperrors.BadRequestError{Errors: map[string]string{"category": "Invalid Category ID"}})
		return
	}

	err = h.categoryUseCase.DeleteCategory(c.Request.Context(), id)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Message(c, "Category Deleted")
}
