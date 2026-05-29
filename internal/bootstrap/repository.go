package bootstrap

import (
	"database/sql"

	"github.com/mafi020/ecom-golang/internal/infrastructure"
)

type Repository struct {
	UserRepo         *infrastructure.PostgresUserRepository
	CategoryRepo     *infrastructure.PostgresCategoryRepository
	RefreshTokenRepo *infrastructure.PostgresRefreshTokenRepository
	ProductRepo      *infrastructure.PostgresProductRepository
	OrderRepo        *infrastructure.PostgresOrderRepository
	OrderItemRepo    *infrastructure.PostgresOrderItemRepository
	CartRepo         *infrastructure.PostgresCartRepository
	CartItemRepo     *infrastructure.PostgresCartItemRepository
	PaymentRepo      *infrastructure.PostgresPaymentRepository
}

func RegisterRepositories(db *sql.DB) *Repository {
	return &Repository{
		UserRepo:         infrastructure.NewPostgresUserRepository(db),
		CategoryRepo:     infrastructure.NewPostgresCategoryRepository(db),
		RefreshTokenRepo: infrastructure.NewPostgresRefreshTokenRepository(db),
		ProductRepo:      infrastructure.NewPostgresProductRepository(db),
		OrderRepo:        infrastructure.NewPostgresOrderRepository(db),
		OrderItemRepo:    infrastructure.NewPostgresOrderItemRepository(db),
		CartRepo:         infrastructure.NewPostgresCartRepository(db),
		CartItemRepo:     infrastructure.NewPostgresCartItemRepository(db),
		PaymentRepo:      infrastructure.NewPostgresPaymentRepository(db),
	}
}
