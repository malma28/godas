package service

import (
	"context"
	"errors"
	"godas/model/domain"
	"godas/model/web"
	"godas/repository"
	"os"

	"github.com/go-playground/validator/v10"
)

type UserService interface {
	Create(web.UserCreateRequest) (web.UserResponse, error)
	FindById(string) (web.UserResponse, error)
	FindAll() ([]web.UserResponse, error)
	Update(string, web.UserUpdateRequest) (web.UserResponse, error)
	Delete(string) error
	Resend(web.EmailVerificationRecreateRequest) error
	Verify(web.EmailVerificationCreateRequest) (web.UserResponse, error)
}

type UserServiceImpl struct {
	userRepository           repository.UserRepository
	emailVerificationService EmailVerificationService
	validate                 *validator.Validate
}

func NewUserService(userRepository repository.UserRepository, emailVerificationService EmailVerificationService, validate *validator.Validate) UserService {
	userService := new(UserServiceImpl)
	userService.userRepository = userRepository
	userService.emailVerificationService = emailVerificationService
	userService.validate = validate

	return userService
}

func (service *UserServiceImpl) Create(request web.UserCreateRequest) (web.UserResponse, error) {
	response := web.UserResponse{}

	if err := service.validate.Struct(request); err != nil {
		return response, ErrBadRequest
	}

	user := domain.User{
		Name:     request.Name,
		Role:     domain.UserRoleClient,
		Email:    request.Email,
		Password: request.Password,
		Verified: false,
	}

	emailVerificationSendChannel := make(chan error)

	go func() {
		_, err := service.emailVerificationService.Create(domain.EmailVerificationSend{
			FromName:     os.Getenv("APP_NAME"),
			FromEmail:    os.Getenv("EMAIL"),
			FromPassword: os.Getenv("EMAIL_PASSWORD"),
			ToEmail:      user.Email,
			Title:        "Email Verification",
			Host:         os.Getenv("EMAIL_HOST"),
			Port:         os.Getenv("EMAIL_PORT"),
		})
		emailVerificationSendChannel <- err
	}()

	user, err := service.userRepository.Insert(context.Background(), user)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateData) {
			return response, ErrDuplicate
		}
		return response, err
	}

	if err := <-emailVerificationSendChannel; err != nil {
		return response, err
	}

	response = web.UserResponse{
		ID:   user.ID,
		Name: user.Name,
	}

	return response, nil
}

func (service *UserServiceImpl) FindById(id string) (web.UserResponse, error) {
	response := web.UserResponse{}

	user, err := service.userRepository.FindById(context.Background(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}

	response = web.UserResponse{
		ID:   user.ID,
		Name: user.Name,
	}

	return response, nil
}

func (service *UserServiceImpl) FindAll() ([]web.UserResponse, error) {
	users, err := service.userRepository.FindAll(context.Background())
	if err != nil {
		return nil, err
	}

	response := []web.UserResponse{}
	for _, user := range users {
		response = append(response, web.UserResponse{
			ID:   user.ID,
			Name: user.Name,
		})
	}

	return response, nil
}

func (service *UserServiceImpl) Update(id string, request web.UserUpdateRequest) (web.UserResponse, error) {
	response := web.UserResponse{}

	if err := service.validate.Struct(request); err != nil {
		return response, ErrBadRequest
	}

	user := domain.User{
		ID:   id,
		Name: request.Name,
	}

	user, err := service.userRepository.Update(context.Background(), user)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}

	response = web.UserResponse{
		ID:   user.ID,
		Name: user.Name,
	}

	return response, nil
}

func (service *UserServiceImpl) Delete(id string) error {
	if err := service.userRepository.Delete(context.Background(), domain.User{ID: id}); err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (service *UserServiceImpl) Resend(request web.EmailVerificationRecreateRequest) error {
	if err := service.validate.Struct(request); err != nil {
		return ErrBadRequest
	}

	_, err := service.emailVerificationService.Recreate(domain.EmailVerificationSend{
		FromName:     os.Getenv("APP_NAME"),
		FromEmail:    os.Getenv("EMAIL"),
		FromPassword: os.Getenv("EMAIL_PASSWORD"),
		ToEmail:      request.Email,
		Title:        "Email Verification",
		Host:         os.Getenv("EMAIL_HOST"),
		Port:         os.Getenv("EMAIL_PORT"),
	})
	if err != nil {
		return err
	}

	return nil
}

func (service *UserServiceImpl) Verify(request web.EmailVerificationCreateRequest) (web.UserResponse, error) {
	response := web.UserResponse{}

	if err := service.validate.Struct(request); err != nil {
		return response, ErrBadRequest
	}

	if err := service.emailVerificationService.Verification(request); err != nil {
		return response, err
	}

	user, err := service.userRepository.FindByEmail(context.Background(), request.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}
	user.Verified = true
	_, err = service.userRepository.Update(context.Background(), user)
	if err != nil {
		if errors.Is(err, repository.ErrNoData) {
			return response, ErrNotFound
		}
		return response, err
	}

	response = web.UserResponse{
		ID:   user.ID,
		Name: user.Name,
	}

	return response, nil
}
