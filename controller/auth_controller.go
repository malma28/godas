package controller

import (
	"errors"
	"godas/model/web"
	"godas/service"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type AuthController interface {
	Signin(*fiber.Ctx) error
	Signup(*fiber.Ctx) error
	EmailVerification(*fiber.Ctx) error
	ResendEmailVerification(*fiber.Ctx) error
}

type AuthControllerImpl struct {
	authService service.AuthService
	userService service.UserService
}

func NewAuthController(authService service.AuthService, userService service.UserService) AuthController {
	authController := new(AuthControllerImpl)
	authController.authService = authService
	authController.userService = userService

	return authController
}

func (controller AuthControllerImpl) Signin(ctx *fiber.Ctx) error {
	request := web.AuthRequest{}

	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	token, err := controller.authService.Signin(request)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrNotFound) || errors.Is(err, service.ErrUnauthorized) {
			statusCode = http.StatusUnauthorized
		}

		return ctx.Status(statusCode).JSON(web.NewFailPayload(statusCode))
	}

	return ctx.JSON(web.Payload{
		Code:    http.StatusOK,
		Status:  http.StatusText(http.StatusOK),
		Success: true,
		Data:    token,
	})
}

func (controller AuthControllerImpl) Signup(ctx *fiber.Ctx) error {
	userCreateRequest := web.UserCreateRequest{}
	if err := ctx.BodyParser(&userCreateRequest); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	response, err := controller.userService.Create(userCreateRequest)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrBadRequest) {
			statusCode = http.StatusBadRequest
		} else if errors.Is(err, service.ErrDuplicate) {
			statusCode = http.StatusConflict
		}
		return ctx.Status(statusCode).JSON(web.NewFailPayload(statusCode))
	}

	return ctx.Status(http.StatusOK).JSON(web.Payload{
		Code:    http.StatusOK,
		Status:  http.StatusText(http.StatusOK),
		Success: true,
		Data:    response,
	})
}

func (controller *AuthControllerImpl) ResendEmailVerification(ctx *fiber.Ctx) error {
	emailRecreateRequest := web.EmailVerificationRecreateRequest{}
	if err := ctx.BodyParser(&emailRecreateRequest); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	if err := controller.userService.Resend(emailRecreateRequest); err != nil {
		log.Println(err.Error())
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrBadRequest) {
			statusCode = http.StatusBadRequest
		} else if errors.Is(err, service.ErrNotFound) {
			statusCode = http.StatusNotFound
		} else if errors.Is(err, service.ErrUnauthorized) {
			statusCode = http.StatusUnauthorized
		} else if errors.Is(err, service.ErrDuplicate) {
			statusCode = http.StatusConflict
		}
		return ctx.Status(statusCode).JSON(web.NewFailPayload(statusCode))
	}

	return ctx.Status(http.StatusOK).JSON(web.Payload{
		Code:    http.StatusOK,
		Status:  http.StatusText(http.StatusOK),
		Success: true,
		Data:    nil,
	})
}

func (controller *AuthControllerImpl) EmailVerification(ctx *fiber.Ctx) error {
	emailVerificationRequest := web.EmailVerificationCreateRequest{}
	if err := ctx.BodyParser(&emailVerificationRequest); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	user, err := controller.userService.Verify(emailVerificationRequest)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrBadRequest) {
			statusCode = http.StatusBadRequest
		} else if errors.Is(err, service.ErrNotFound) {
			statusCode = http.StatusNotFound
		} else if errors.Is(err, service.ErrUnauthorized) {
			statusCode = http.StatusUnauthorized
		}
		return ctx.Status(statusCode).JSON(web.NewFailPayload(statusCode))
	}

	return ctx.Status(http.StatusOK).JSON(web.Payload{
		Code:    http.StatusOK,
		Status:  http.StatusText(http.StatusOK),
		Success: true,
		Data:    user,
	})
}
