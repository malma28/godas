package controller

import (
	"errors"
	"godas/model/web"
	"godas/service"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type StackController interface {
	Create(ctx *fiber.Ctx) error
	FindById(ctx *fiber.Ctx) error
	FindAll(ctx *fiber.Ctx) error
	Push(ctx *fiber.Ctx) error
	Pop(ctx *fiber.Ctx) error
}

type StackControllerImpl struct {
	stackService service.StackService
}

func NewStackController(stackService service.StackService) StackController {
	controller := new(StackControllerImpl)
	controller.stackService = stackService

	return controller
}

func (controller *StackControllerImpl) Create(ctx *fiber.Ctx) error {
	authResponse, isAuthResponse := ctx.UserContext().Value("response").(web.AuthResponse)
	if !isAuthResponse {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	response, err := controller.stackService.Create(authResponse.ID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrDuplicate) {
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

func (controller *StackControllerImpl) FindById(ctx *fiber.Ctx) error {
	authResponse, isAuthResponse := ctx.UserContext().Value("response").(web.AuthResponse)
	if !isAuthResponse {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	stack, err := controller.stackService.FindByIdFromOwner(id, authResponse.ID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrNotFound) {
			statusCode = http.StatusNotFound
		}
		return ctx.Status(statusCode).JSON(web.NewFailPayload(statusCode))
	}

	return ctx.JSON(web.Payload{
		Code:    http.StatusOK,
		Status:  http.StatusText(http.StatusOK),
		Success: true,
		Data:    stack,
	})
}

func (controller *StackControllerImpl) FindAll(ctx *fiber.Ctx) error {
	authResponse, isAuthResponse := ctx.UserContext().Value("response").(web.AuthResponse)
	if !isAuthResponse {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	stacks, err := controller.stackService.FindAllFromOwner(authResponse.ID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrNotFound) {
			statusCode = http.StatusNotFound
		}
		return ctx.Status(statusCode).JSON(web.NewFailPayload(statusCode))
	}

	return ctx.JSON(web.Payload{
		Code:    http.StatusOK,
		Status:  http.StatusText(http.StatusOK),
		Success: true,
		Data:    stacks,
	})
}

func (controller *StackControllerImpl) Push(ctx *fiber.Ctx) error {
	authResponse, isAuthResponse := ctx.UserContext().Value("response").(web.AuthResponse)
	if !isAuthResponse {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	request := web.ItemRequest{}
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	response, err := controller.stackService.PushFromOwner(id, authResponse.ID, request)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrBadRequest) {
			statusCode = http.StatusBadRequest
		} else if errors.Is(err, service.ErrNotFound) {
			statusCode = http.StatusNotFound
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

func (controller *StackControllerImpl) Pop(ctx *fiber.Ctx) error {
	authResponse, isAuthResponse := ctx.UserContext().Value("response").(web.AuthResponse)
	if !isAuthResponse {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(web.NewFailPayload(http.StatusBadRequest))
	}

	response, err := controller.stackService.PopFromOwner(id, authResponse.ID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrNotFound) {
			statusCode = http.StatusNotFound
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
