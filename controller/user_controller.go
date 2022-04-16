package controller

import (
	"errors"
	"godas/model/domain"
	"godas/model/web"
	"godas/service"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type UserController interface {
	Create(*fiber.Ctx) error
	FindById(*fiber.Ctx) error
	FindAll(*fiber.Ctx) error
	Update(*fiber.Ctx) error
	Delete(*fiber.Ctx) error
}

type UserControllerImpl struct {
	service service.UserService
}

func NewUserController(service service.UserService) UserController {
	userController := new(UserControllerImpl)
	userController.service = service

	return userController
}

func (controller *UserControllerImpl) Create(ctx *fiber.Ctx) error {
	authResponse, isAuthResponse := ctx.UserContext().Value("response").(web.AuthResponse)
	if !isAuthResponse {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	if authResponse.Role != domain.UserRoleAdmin {
		return ctx.Status(http.StatusUnauthorized).JSON(web.NewFailPayload(http.StatusUnauthorized))
	}

	userCreateRequest := web.UserCreateRequest{}
	if err := ctx.BodyParser(&userCreateRequest); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	response, err := controller.service.Create(userCreateRequest)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrBadRequest) {
			statusCode = http.StatusBadRequest
		} else if errors.Is(err, service.ErrDuplicate) {
			statusCode = http.StatusConflict
		}
		return ctx.Status(statusCode).JSON(web.NewFailPayload(statusCode))
	}

	return ctx.JSON(web.Payload{
		Code:    http.StatusOK,
		Status:  http.StatusText(http.StatusOK),
		Success: true,
		Data:    response,
	})
}

func (controller *UserControllerImpl) FindById(ctx *fiber.Ctx) error {
	authResponse, isAuthResponse := ctx.UserContext().Value("response").(web.AuthResponse)
	if !isAuthResponse {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	} else if id == "me" {
		id = authResponse.ID
	}

	user, err := controller.service.FindById(id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrNotFound) {
			statusCode = http.StatusNotFound
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

func (controller *UserControllerImpl) FindAll(ctx *fiber.Ctx) error {
	users, err := controller.service.FindAll()
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(web.NewFailPayload(http.StatusInternalServerError))
	}

	return ctx.Status(http.StatusOK).JSON(web.Payload{
		Code:    http.StatusOK,
		Status:  http.StatusText(http.StatusOK),
		Success: true,
		Data:    users,
	})
}

func (controller *UserControllerImpl) Update(ctx *fiber.Ctx) error {
	authResponse, isAuthResponse := ctx.UserContext().Value("response").(web.AuthResponse)
	if !isAuthResponse {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	} else if id == "me" {
		id = authResponse.ID
	} else {
		if authResponse.Role != domain.UserRoleAdmin {
			return ctx.Status(http.StatusUnauthorized).JSON(web.NewFailPayload(http.StatusUnauthorized))
		}
	}

	request := web.UserUpdateRequest{}
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	user, err := controller.service.Update(id, request)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrNotFound) {
			statusCode = http.StatusNotFound
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

func (controller *UserControllerImpl) Delete(ctx *fiber.Ctx) error {
	authResponse, isAuthResponse := ctx.UserContext().Value("response").(web.AuthResponse)
	if !isAuthResponse {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	} else if id == "me" {
		id = authResponse.ID
	} else {
		if authResponse.Role != domain.UserRoleAdmin {
			return ctx.Status(http.StatusUnauthorized).JSON(web.NewFailPayload(http.StatusUnauthorized))
		}
	}

	if err := controller.service.Delete(id); err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrNotFound) {
			statusCode = http.StatusNotFound
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
