package usecase

import (
	"context"
	"fmt"

	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/entity"
)

type orderRepo interface {
	CreateOrder(ctx context.Context, order *entity.Order) error
	GetOrderByID(ctx context.Context, id, userID int64) (*entity.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int64, params entity.GetOrdersParams) ([]entity.Order, int, error)
}

type orderItemRepo interface {
	CreateOrderItems(ctx context.Context, orderID int64, items []entity.OrderItem) ([]entity.OrderItem, error)
}

type orderProductRepo interface {
	BatchUpdate(ctx context.Context, updates map[int64]*entity.UpdateProductInput) error
	GetByIDs(ctx context.Context, ids []int64) ([]entity.Product, error)
}

type orderCartRepo interface {
	GetCartByUserID(ctx context.Context, userID int64) (*entity.Cart, error)
}

type orderCartItemRepo interface {
	ClearCart(ctx context.Context, cartID int64) error
}

type transactionManager interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}

type OrderUseCase struct {
	orderRepo     orderRepo
	orderItemRepo orderItemRepo
	productRepo   orderProductRepo
	cartRepo      orderCartRepo
	cartItemRepo  orderCartItemRepo
	tm            transactionManager
}

func NewOrderUseCase(orderRepo orderRepo, orderItemRepo orderItemRepo, productRepo orderProductRepo, cartRepo orderCartRepo, cartItemRepo orderCartItemRepo, tm transactionManager) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		productRepo:   productRepo,
		cartRepo:      cartRepo,
		cartItemRepo:  cartItemRepo,
		tm:            tm,
	}
}

func (uc *OrderUseCase) Checkout(ctx context.Context, userID int64) (*entity.Order, error) {
	cart, err := uc.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(cart.Items) == 0 {
		return nil, &apperrors.ValidationError{Errors: map[string]string{
			"cart": "cart is empty",
		}}
	}

	result, err := uc.prepareOrderItems(ctx, cart.Items)
	if err != nil {
		return nil, err
	}

	var order *entity.Order

	err = uc.tm.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.productRepo.BatchUpdate(txCtx, result.stockUpdates); err != nil {
			return err
		}

		order = &entity.Order{
			UserID:     userID,
			Status:     entity.OrderStatusPending,
			TotalPrice: result.totalPrice,
			OrderItems: result.items,
		}

		if err := uc.orderRepo.CreateOrder(txCtx, order); err != nil {
			return err
		}

		updatedItems, err := uc.orderItemRepo.CreateOrderItems(txCtx, order.ID, result.items)
		if err != nil {
			return err
		}
		order.OrderItems = updatedItems

		// 6. Clear cart after successful order
		if err := uc.cartItemRepo.ClearCart(txCtx, cart.ID); err != nil {
			return err
		}

		return nil
	})

	return order, err
}

type checkoutResult struct {
	totalPrice   float64
	items        []entity.OrderItem
	stockUpdates map[int64]*entity.UpdateProductInput
}

func (uc *OrderUseCase) prepareOrderItems(ctx context.Context, cartItems []entity.CartItem) (*checkoutResult, error) {
	productIDs := make([]int64, len(cartItems))
	for i, item := range cartItems {
		productIDs[i] = item.ProductID
	}

	products, err := uc.productRepo.GetByIDs(ctx, productIDs)
	if err != nil {
		return nil, err
	}

	productMap := make(map[int64]entity.Product, len(products))
	for _, p := range products {
		productMap[p.ID] = p
	}

	var totalPrice float64
	items := make([]entity.OrderItem, len(cartItems))
	stockUpdates := make(map[int64]*entity.UpdateProductInput, len(cartItems))

	for i, cartItem := range cartItems {
		product, ok := productMap[cartItem.ProductID]
		if !ok {
			return nil, &apperrors.NotFoundError{Resource: "product"}
		}

		if product.Stock < cartItem.Quantity {
			return nil, &apperrors.ValidationError{Errors: map[string]string{
				"stock": fmt.Sprintf("insufficient stock for product %d: have %d, want %d",
					product.ID, product.Stock, cartItem.Quantity),
			}}
		}

		items[i] = entity.OrderItem{
			ProductID: cartItem.ProductID,
			Quantity:  cartItem.Quantity,
			Price:     product.Price,
		}
		totalPrice += product.Price * float64(cartItem.Quantity)

		newStock := product.Stock - cartItem.Quantity
		stockUpdates[product.ID] = &entity.UpdateProductInput{Stock: &newStock}
	}

	return &checkoutResult{
		totalPrice:   totalPrice,
		items:        items,
		stockUpdates: stockUpdates,
	}, nil
}

func (uc *OrderUseCase) GetOrderByID(ctx context.Context, id, userID int64) (*entity.Order, error) {
	return uc.orderRepo.GetOrderByID(ctx, id, userID)
}

func (uc *OrderUseCase) GetOrdersByUserID(ctx context.Context, userID int64, params entity.GetOrdersParams) ([]entity.Order, int, error) {
	return uc.orderRepo.GetOrdersByUserID(ctx, userID, params)
}
