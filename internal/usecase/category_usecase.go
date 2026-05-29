package usecase

import (
	"context"
	"fmt"

	"github.com/mafi020/ecom-golang/internal/delivery/http/utils"
	"github.com/mafi020/ecom-golang/internal/entity"
)

type categoryRepository interface {
	CreateCategory(ctx context.Context, category *entity.Category) error
	GetCategoryByID(ctx context.Context, id int64) (*entity.Category, error)
	GetCategoryByIDWithProducts(ctx context.Context, id int64) (*entity.Category, error)
	GetAllCategories(ctx context.Context, params entity.GetCategoriesParams) ([]entity.Category, int, error)
	UpdateCategory(ctx context.Context, category *entity.Category) error
	DeleteCategory(ctx context.Context, id int64) error
	IsCycle(ctx context.Context, categoryID int64, parentID int64) (bool, error)
}

type CategoryUseCase struct {
	repo categoryRepository
}

func NewCategoryUseCase(repo categoryRepository) *CategoryUseCase {
	return &CategoryUseCase{repo: repo}
}

func (uc *CategoryUseCase) CreateCategory(ctx context.Context, name string, parentID *int64) (*entity.Category, error) {
	slug := utils.GenerateSlug(name)

	category := &entity.Category{
		Name:     name,
		Slug:     slug,
		ParentID: parentID,
	}
	err := uc.repo.CreateCategory(ctx, category)

	if err != nil {
		return nil, err
	}
	return category, nil
}

func (uc *CategoryUseCase) GetCategoryByID(ctx context.Context, id int64) (*entity.Category, error) {
	return uc.repo.GetCategoryByID(ctx, id)
}

func (uc *CategoryUseCase) GetCategoryByIDWithProducts(ctx context.Context, id int64) (*entity.Category, error) {
	return uc.repo.GetCategoryByIDWithProducts(ctx, id)
}

func (uc *CategoryUseCase) GetAllCategories(ctx context.Context, params entity.GetCategoriesParams) ([]entity.Category, int, error) {
	return uc.repo.GetAllCategories(ctx, params)
}

func (uc *CategoryUseCase) UpdateCategory(ctx context.Context, category *entity.Category) error {
	// 1. Self-parent check
	if category.ParentID != nil && *category.ParentID == category.ID {
		return fmt.Errorf("a category cannot be its own parent")
	}

	// 2. Cycle detection: only when a parent is being assigned
	if category.ParentID != nil {
		hasCycle, err := uc.repo.IsCycle(ctx, category.ID, *category.ParentID)
		if err != nil {
			return fmt.Errorf("failed to check category cycle: %w", err)
		}
		if hasCycle {
			return fmt.Errorf("cannot update: category %d would create a circular reference", category.ID)
		}
	}

	// 3. Persist to DB
	return uc.repo.UpdateCategory(ctx, category)
}

func (uc *CategoryUseCase) DeleteCategory(ctx context.Context, id int64) error {
	return uc.repo.DeleteCategory(ctx, id)
}
