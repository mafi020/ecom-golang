package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/mafi020/ecom-golang/internal/apperrors"
	"github.com/mafi020/ecom-golang/internal/delivery/http/request"
	"github.com/mafi020/ecom-golang/internal/delivery/http/utils"
	"github.com/mafi020/ecom-golang/internal/entity"
	"github.com/mafi020/ecom-golang/internal/response"
)

type authUseCase interface {
	Register(ctx context.Context, user *entity.User) error
	Login(ctx context.Context, email string, password string) (user *entity.User, accessToken string, refreshToken string, err error)
	RefreshAccessToken(ctx context.Context, rawRefreshToken string) (string, error)
	Logout(ctx context.Context, rawRefreshToken string) error
}

type AuthHandler struct {
	authUseCase authUseCase
}

func NewAuthHandler(uc authUseCase) *AuthHandler {
	return &AuthHandler{authUseCase: uc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, apperrors.ParseValidationError(err))
		return
	}

	user := entity.NewUser(req.Name, req.Email, req.Password, req.Role)

	if err := h.authUseCase.Register(c.Request.Context(), user); err != nil {
		utils.HandleError(c, err)
		return
	}
	response.Success(c, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, apperrors.ParseValidationError(err))
		return
	}

	user, accessToken, refreshToken, err := h.authUseCase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, apperrors.ParseValidationError(err))
		return
	}

	accessToken, err := h.authUseCase.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Success(c, gin.H{"access_token": accessToken})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, apperrors.ParseValidationError(err))
		return
	}

	if err := h.authUseCase.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		utils.HandleError(c, err)
		return
	}

	response.Message(c, "logged out successfully")
}
