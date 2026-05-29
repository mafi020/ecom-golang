package entity

import (
	"time"

	"github.com/mafi020/ecom-golang/internal/utils"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleCustomer Role = "customer"
)

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Exclude from JSON responses
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	RefreshTokens []RefreshToken `json:"refresh_tokens,omitempty"`
}

func NewUser(name, email, password string, role Role) *User {
	if role == "" {
		role = RoleCustomer
	}
	return &User{
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
	}
}

type GetUsersParams struct {
	utils.QueryParams
	Role string `form:"role"`
}
