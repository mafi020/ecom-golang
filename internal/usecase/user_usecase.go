package usecase

import (
	"context"

	"github.com/mafi020/ecom-golang/internal/entity"
)

type UserInterface interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	GetAll(ctx context.Context, params entity.GetUsersParams) ([]entity.User, int, error)
	Delete(ctx context.Context, id int64) error
}

type UserUseCase struct {
	repo UserInterface
}

func NewUserUsecase(repo UserInterface) *UserUseCase {
	return &UserUseCase{repo: repo}
}

func (uc *UserUseCase) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	return uc.repo.GetByID(ctx, id)
}

// GetUsers retrieves all users.
func (uc *UserUseCase) GetUsers(ctx context.Context, params entity.GetUsersParams) ([]entity.User, int, error) {
	return uc.repo.GetAll(ctx, params)
}

func (uc *UserUseCase) DeleteUser(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}
