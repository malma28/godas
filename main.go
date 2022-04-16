package main

import (
	"godas/app"
	"godas/controller"
	"godas/middleware"
	"godas/repository"
	"godas/service"
	"os"
)

func main() {
	test := false
	if len(os.Args) > 1 && os.Args[1] == "--test" {
		test = true
	}

	mainApp := app.New(test)

	emailVerificationRepository := repository.NewEmailVerificationRepository(mainApp.DB)
	emailVerificationService := service.NewEmailVerificationService(emailVerificationRepository, mainApp.Validate)

	userRepository := repository.NewUserRepository(mainApp.DB, mainApp.SnowflakeNode)
	userService := service.NewUserService(userRepository, emailVerificationService, mainApp.Validate)
	userController := controller.NewUserController(userService)

	authService := service.NewAuthService(userRepository, mainApp.JWTProvider)
	authController := controller.NewAuthController(authService, userService)
	authMiddleware := middleware.NewAuthMiddleware(authService)

	stackRepository := repository.NewStackRepository(mainApp.DB, mainApp.SnowflakeNode)
	stackService := service.NewStackService(stackRepository, userRepository, mainApp.Validate)
	stackController := controller.NewStackController(stackService)

	docsController := controller.NewDocsController()

	mainApp.SetupRouter(userController, authController, stackController, docsController, authMiddleware)

	mainApp.Run()
}
