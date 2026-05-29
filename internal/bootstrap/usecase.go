package bootstrap

import (
	"github.com/mafi020/ecom-golang/config"
	"github.com/mafi020/ecom-golang/internal/infrastructure"
	"github.com/mafi020/ecom-golang/internal/usecase"
)

type Usecases struct {
	AuthUC         *usecase.AuthUseCase
	UserUC         *usecase.UserUseCase
	RefreshTokenUC *usecase.AuthUseCase
	CategoryUC     *usecase.CategoryUseCase
	ProductUC      *usecase.ProductUseCase
	CartUC         *usecase.CartUseCase
	OrderUC        *usecase.OrderUseCase
	PaymentUC      *usecase.PaymentUseCase
}

func RegisterUsecases(repo *Repository, tm *infrastructure.Transactor, cfg *config.Config) *Usecases {
	return &Usecases{
		AuthUC:         usecase.NewAuthUsecase(repo.UserRepo, repo.RefreshTokenRepo, cfg),
		UserUC:         usecase.NewUserUsecase(repo.UserRepo),
		RefreshTokenUC: usecase.NewAuthUsecase(repo.UserRepo, repo.RefreshTokenRepo, cfg),
		CategoryUC:     usecase.NewCategoryUseCase(repo.CategoryRepo),
		ProductUC:      usecase.NewProductUseCase(repo.ProductRepo, repo.CategoryRepo),
		CartUC:         usecase.NewCartUsecase(repo.CartRepo, repo.CartItemRepo, repo.ProductRepo),
		OrderUC:        usecase.NewOrderUseCase(repo.OrderRepo, repo.OrderItemRepo, repo.ProductRepo, repo.CartRepo, repo.CartItemRepo, tm),
		PaymentUC:      usecase.NewPaymentUseCase(repo.PaymentRepo, repo.OrderRepo, tm),
	}
}
