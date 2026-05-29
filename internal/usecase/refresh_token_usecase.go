package usecase

import (
	"context"

	"github.com/mafi020/ecom-golang/internal/entity"
)

type refreshTokenInterface interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error)
	DeleteByToken(ctx context.Context, token string) error
	DeleteAllByUserID(ctx context.Context, userID int64) error
}

type RefreshTokenUsecase struct {
	repo refreshTokenInterface
}

func NewRefreshTokenUsecase(repo refreshTokenInterface) *RefreshTokenUsecase {
	return &RefreshTokenUsecase{
		repo: repo,
	}
}

func (uc *RefreshTokenUsecase) Create(ctx context.Context, token *entity.RefreshToken) error {
	return uc.repo.Create(ctx, token)
}

func (uc *RefreshTokenUsecase) GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	return uc.repo.GetByToken(ctx, token)
}

func (uc *RefreshTokenUsecase) DeleteByToken(ctx context.Context, token string) error {
	return uc.repo.DeleteByToken(ctx, token)
}

func (uc *RefreshTokenUsecase) DeleteAllByUserID(ctx context.Context, userID int64) error {
	return uc.repo.DeleteAllByUserID(ctx, userID)
}
