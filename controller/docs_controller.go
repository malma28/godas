package controller

import "github.com/gofiber/fiber/v2"

type DocsController interface {
	HTML(ctx *fiber.Ctx) error
}

func NewDocsController() DocsController {
	docsController := new(DocsControllerImpl)

	return docsController
}

type DocsControllerImpl struct {
}

func (controller *DocsControllerImpl) HTML(ctx *fiber.Ctx) error {
	ctx.Context().SetContentType("text/html")
	return ctx.SendFile("./docs/docs.html")
}
