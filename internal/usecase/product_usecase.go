package usecase

import (
	"context"

	"github.com/mafi020/ecom-golang/internal/entity"
)

type productInterface interface {
	Create(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, id int64) (*entity.Product, error)
	Update(ctx context.Context, id int64, input *entity.UpdateProductInput) (*entity.Product, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, params entity.GetProductsParams) ([]entity.Product, int, error)
}

type productCategoryInterface interface {
	GetCategoryByID(ctx context.Context, id int64) (*entity.Category, error)
}

type ProductUseCase struct {
	productRepo  productInterface
	categoryRepo productCategoryInterface
}

func NewProductUseCase(productRepo productInterface, categoryRepo productCategoryInterface) *ProductUseCase {
	return &ProductUseCase{productRepo: productRepo, categoryRepo: categoryRepo}
}

func (uc *ProductUseCase) Create(ctx context.Context, product *entity.Product) error {
	_, err := uc.categoryRepo.GetCategoryByID(ctx, product.CategoryID)
	if err != nil {
		return err
	}
	return uc.productRepo.Create(ctx, product)
}

func (uc *ProductUseCase) GetByID(ctx context.Context, id int64) (*entity.Product, error) {
	return uc.productRepo.GetByID(ctx, id)
}

func (uc *ProductUseCase) Update(ctx context.Context, id int64, input *entity.UpdateProductInput) (*entity.Product, error) {
	return uc.productRepo.Update(ctx, id, input)
}

func (uc *ProductUseCase) Delete(ctx context.Context, id int64) error {
	return uc.productRepo.Delete(ctx, id)
}

func (uc *ProductUseCase) List(ctx context.Context, params entity.GetProductsParams) ([]entity.Product, int, error) {
	return uc.productRepo.List(ctx, params)
}
