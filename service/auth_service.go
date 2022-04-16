package service

import (
	"context"
	"errors"
	"godas/model/web"
	"godas/repository"
	"godas/secure"
)

type AuthService interface {
	Signin(web.AuthRequest) (string, error)
	Validate(string) (web.AuthResponse, error)
}

type AuthServiceImpl struct {
	userRepository repository.UserRepository
	jwtProvider    *secure.JWTProvider
}

func NewAuthService(userRepository repository.UserRepository, jwtProvider *secure.JWTProvider) AuthService {
	authService := new(AuthServiceImpl)
	authService.userRepository = userRepository
	authService.jwtProvider = jwtProvider

	return authService
}

func (service *AuthServiceImpl) Signin(request web.AuthRequest) (string, error) {
	user, err := service.userRepository.FindByEmail(context.Background(), request.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return "", ErrNotFound
		}
		return "", err
	}

	if user.Password != request.Password || !user.Verified {
		return "", ErrUnauthorized
	}

	return service.jwtProvider.Token(web.JwtClaims{
		UserID: user.ID,
		Role:   user.Role,
	})
}

func (service *AuthServiceImpl) Validate(tokenString string) (web.AuthResponse, error) {
	response := web.AuthResponse{}

	claims, err := service.jwtProvider.Validate(tokenString)
	if err != nil {
		return response, err
	}

	user, err := service.userRepository.FindById(context.Background(), claims.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}
	if !user.Verified {
		return response, ErrUnauthorized
	}

	response = web.AuthResponse{
		ID:   claims.UserID,
		Role: claims.Role,
	}

	return response, nil
}
