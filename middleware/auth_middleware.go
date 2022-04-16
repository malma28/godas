package middleware

import (
	"context"
	"errors"
	"godas/model/web"
	"godas/service"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware struct {
	authService service.AuthService
}

func NewAuthMiddleware(authService service.AuthService) *AuthMiddleware {
	authMiddleware := new(AuthMiddleware)
	authMiddleware.authService = authService

	return authMiddleware
}

func (middleware *AuthMiddleware) Use(prefixs ...string) func(ctx *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		path := ctx.Path()
		accept := false
		for _, prefix := range prefixs {
			if strings.HasPrefix(path, prefix) {
				accept = true
			}
		}

		if !accept {
			return ctx.Next()
		}

		authorization, isExist := ctx.GetReqHeaders()["Authorization"]
		if !isExist {
			return ctx.Status(http.StatusUnauthorized).JSON(web.NewFailPayload(http.StatusUnauthorized))
		}

		if strings.HasPrefix(authorization, "Bearer ") {
			authorization = strings.TrimPrefix(authorization, "Bearer ")
		} else {
			return ctx.Status(http.StatusUnauthorized).JSON(web.NewFailPayload(http.StatusUnauthorized))
		}

		response, err := middleware.authService.Validate(authorization)
		if err != nil {
			statusCode := http.StatusBadRequest
			if errors.Is(err, service.ErrNotFound) || errors.Is(err, service.ErrUnauthorized) {
				statusCode = http.StatusUnauthorized
			}
			return ctx.Status(statusCode).JSON(web.NewFailPayload(statusCode))
		}

		ctx.SetUserContext(context.WithValue(context.Background(), "response", response))
		return ctx.Next()
	}
}
