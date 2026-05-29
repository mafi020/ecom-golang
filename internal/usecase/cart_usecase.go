package usecase

import (
	"context"

	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/entity"
)

type cartRepo interface {
	GetOrCreateCart(ctx context.Context, userID int64) (*entity.Cart, error)
	GetCartByUserID(ctx context.Context, userID int64) (*entity.Cart, error)
}

type cartItemRepo interface {
	AddItem(ctx context.Context, cartID int64, item *entity.CartItem) (*entity.CartItem, error)
	UpdateItemQuantity(ctx context.Context, cartID, productID int64, quantity int) (*entity.CartItem, error)
	RemoveItem(ctx context.Context, cartID, productID int64) error
	ClearCart(ctx context.Context, cartID int64) error
}

type cartProductRepo interface {
	GetByID(ctx context.Context, id int64) (*entity.Product, error)
}

type CartUseCase struct {
	cartRepo        cartRepo
	cartItemRepo    cartItemRepo
	cartProductRepo cartProductRepo
}

func NewCartUsecase(cartRepo cartRepo, cartItemRepo cartItemRepo, cartProductRepo cartProductRepo) *CartUseCase {
	return &CartUseCase{cartRepo: cartRepo, cartItemRepo: cartItemRepo, cartProductRepo: cartProductRepo}
}

func (uc *CartUseCase) GetCart(ctx context.Context, userID int64) (*entity.Cart, error) {
	return uc.cartRepo.GetCartByUserID(ctx, userID)
}

func (uc *CartUseCase) AddItem(ctx context.Context, userID, productID int64, quantity int) (*entity.Cart, error) {
	cart, product, totalQuantityRequested, err := uc.prepareAndValidateStock(ctx, userID, productID, quantity, false)
	if err != nil {
		return nil, err
	}

	_, err = uc.cartItemRepo.AddItem(ctx, cart.ID, &entity.CartItem{
		ProductID: productID,
		Quantity:  totalQuantityRequested,
		Price:     product.Price,
	})

	if err != nil {
		return nil, err
	}

	return uc.cartRepo.GetCartByUserID(ctx, userID)
}

func (uc *CartUseCase) UpdateItem(ctx context.Context, userID, productID int64, quantity int) (*entity.Cart, error) {
	cart, _, totalQuantityRequested, err := uc.prepareAndValidateStock(ctx, userID, productID, quantity, true)
	if err != nil {
		return nil, err
	}

	_, err = uc.cartItemRepo.UpdateItemQuantity(ctx, cart.ID, productID, totalQuantityRequested)
	if err != nil {
		return nil, err
	}

	return uc.cartRepo.GetCartByUserID(ctx, userID)
}

func (uc *CartUseCase) RemoveItem(ctx context.Context, userID, productID int64) (*entity.Cart, error) {
	cart, err := uc.cartRepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := uc.cartItemRepo.RemoveItem(ctx, cart.ID, productID); err != nil {
		return nil, err
	}

	return uc.cartRepo.GetCartByUserID(ctx, userID)
}

func (uc *CartUseCase) ClearCart(ctx context.Context, userID int64) error {
	cart, err := uc.cartRepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return err
	}
	return uc.cartItemRepo.ClearCart(ctx, cart.ID)
}

func (uc *CartUseCase) prepareAndValidateStock(
	ctx context.Context,
	userID, productID int64,
	inputQuantity int,
	isAbsoluteUpdate bool,
) (*entity.Cart, *entity.Product, int, error) {
	product, err := uc.cartProductRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, nil, 0, err
	}

	cart, err := uc.cartRepo.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, nil, 0, err
	}

	existingQuantityInCart := 0
	for _, item := range cart.Items {
		if item.ProductID == productID {
			existingQuantityInCart = item.Quantity
			break
		}
	}

	// Determine requested amount depending on action type
	totalQuantityRequested := inputQuantity
	if !isAbsoluteUpdate {
		totalQuantityRequested = existingQuantityInCart + inputQuantity
	}

	if product.Stock < totalQuantityRequested {
		return nil, nil, 0, &apperrors.BadRequestError{
			Errors: map[string]string{"quantity": "insufficient stock"},
		}
	}

	return cart, product, totalQuantityRequested, nil
}
