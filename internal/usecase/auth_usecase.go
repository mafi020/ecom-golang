package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mafi020/ecom-golang/config"
	"github.com/mafi020/ecom-golang/internal/entity"
	"github.com/mafi020/ecom-golang/internal/utils"
)

type authUserInterface interface {
	Create(ctx context.Context, user *entity.User) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByID(ctx context.Context, id int64) (*entity.User, error)
}

type authRefreshTokenInterface interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*entity.RefreshToken, error)
	DeleteByToken(ctx context.Context, token string) error
}

type AuthUseCase struct {
	userRepo         authUserInterface
	refreshTokenRepo authRefreshTokenInterface
	config           *config.Config
}

func NewAuthUsecase(userRepo authUserInterface, refreshTokenRepo authRefreshTokenInterface, config *config.Config) *AuthUseCase {
	return &AuthUseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		config:           config,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, user *entity.User) error {
	hashed, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashed
	return uc.userRepo.Create(ctx, user)
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (user *entity.User, accessToken string, refreshToken string, err error) {

	storedUser, err := uc.userRepo.GetByEmail(ctx, email)

	if err != nil {
		return nil, "", "", fmt.Errorf("invalid credentials %w", err)
	}

	// 2. Compare password
	if err := utils.ComparePassword(storedUser.Password, password); err != nil {
		return nil, "", "", fmt.Errorf("invalid credentials %w", err)
	}

	// 3. Generate access token
	accessToken, err = uc.generateAccessToken(storedUser)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token %w", err)
	}

	// 4. Generate and store refresh token
	refreshToken, err = uc.generateAndStoreRefreshToken(ctx, storedUser.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token %w", err)
	}

	return storedUser, accessToken, refreshToken, nil
}

func (uc *AuthUseCase) RefreshAccessToken(ctx context.Context, rawRefreshToken string) (string, error) {

	storedToken, err := uc.refreshTokenRepo.GetByToken(ctx, rawRefreshToken)
	if err != nil {
		return "", err
	}

	// 2. Check expiry
	if time.Now().After(storedToken.ExpiresAt) {
		uc.refreshTokenRepo.DeleteByToken(ctx, rawRefreshToken)
		return "", fmt.Errorf("refresh token expired %w", err)
	}

	// 3. Get user
	user, err := uc.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return "", err
	}

	// 4. Generate new access token
	return uc.generateAccessToken(user)
}

func (uc *AuthUseCase) Logout(ctx context.Context, rawRefreshToken string) error {
	return uc.refreshTokenRepo.DeleteByToken(ctx, rawRefreshToken)
}

// ── helpers ──────────────────────────────────────────────

func (uc *AuthUseCase) generateAccessToken(user *entity.User) (string, error) {
	secret := uc.config.JWT.Secret
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(uc.config.JWTAccessExpiry()).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (uc *AuthUseCase) generateAndStoreRefreshToken(ctx context.Context, userID int64) (string, error) {
	secret := uc.config.JWT.Secret
	expiresAt := time.Now().Add(uc.config.JWTRefreshExpiry())

	claims := jwt.MapClaims{
		"sub": userID,
		"exp": expiresAt.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("invalid token %w", err)
	}

	// Store in DB
	refreshToken := &entity.RefreshToken{
		UserID:    userID,
		Token:     tokenStr,
		ExpiresAt: expiresAt,
	}

	if err := uc.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return "", err
	}

	return tokenStr, nil
}
