package web

import (
	"godas/model/domain"

	"github.com/golang-jwt/jwt/v4"
)

type JwtClaims struct {
	UserID string          `json:"id"`
	Role   domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

type AuthRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type AuthResponse struct {
	ID   string          `json:"id"`
	Role domain.UserRole `json:"role"`
}
