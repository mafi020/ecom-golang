package request

import "github.com/mafi020/ecom-golang/internal/entity"

type RegisterRequest struct {
	Name            string      `json:"name" binding:"required"`
	Email           string      `json:"email" binding:"required,email"`
	Password        string      `json:"password" binding:"required,min=8"`
	ConfirmPassword string      `json:"confirm_password" binding:"required,eqfield=Password"`
	Role            entity.Role `json:"role" binding:"omitempty,oneof=admin customer"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
