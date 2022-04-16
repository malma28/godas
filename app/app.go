package app

import (
	"context"
	"fmt"
	"godas/controller"
	"godas/middleware"
	"godas/secure"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type App struct {
	client        *mongo.Client
	Core          *fiber.App
	Ctx           context.Context
	DB            *mongo.Database
	SnowflakeNode *snowflake.Node
	Validate      *validator.Validate
	JWTProvider   *secure.JWTProvider
}

func New(test bool) *App {
	var err error

	dotenvFile := "production.env"
	if test {
		dotenvFile = "test.env"
	}
	if err := godotenv.Load(dotenvFile); err != nil {
		panic(err)
	}

	app := new(App)
	app.Core = fiber.New(fiber.Config{
		CaseSensitive:     true,
		ReduceMemoryUsage: true,
		StrictRouting:     true,
	})
	app.Ctx = context.Background()
	app.client, err = mongo.Connect(app.Ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		panic(err)
	}
	app.DB = app.client.Database(os.Getenv("DATABASE_NAME"))
	app.SnowflakeNode, err = snowflake.NewNode(8)
	if err != nil {
		panic(err)
	}
	app.Validate = validator.New()
	app.JWTProvider = secure.NewJWTProvider(secure.DefaultJWTExpiration, os.Getenv("APP_NAME"), os.Getenv("JWT_SIGNATURE_KEY"))
	rand.Seed(time.Now().UnixNano())

	return app
}

func (app *App) SetupRouter(
	userController controller.UserController,
	authController controller.AuthController,
	stackController controller.StackController,
	docsController controller.DocsController,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Auth Controller
	app.Core.Post("/signin", authController.Signin)
	app.Core.Post("/signup", authController.Signup)
	app.Core.Post("/verification", authController.EmailVerification)
	app.Core.Post("/resend", authController.ResendEmailVerification)
	app.Core.Use(authMiddleware.Use("/users", "/stacks"))

	// User Controller
	usersGroup := app.Core.Group("/users")
	usersGroup.Post("", userController.Create)
	usersGroup.Get("/:id", userController.FindById)
	usersGroup.Get("", userController.FindAll)
	usersGroup.Put("/:id", userController.Update)
	usersGroup.Delete("/:id", userController.Delete)

	// Stack Controller
	stacksGroup := app.Core.Group("/stacks")
	stacksGroup.Post("", stackController.Create)
	stacksGroup.Get("/:id", stackController.FindById)
	stacksGroup.Get("", stackController.FindAll)
	stacksGroup.Post("/:id", stackController.Push)
	stacksGroup.Delete("/:id", stackController.Pop)

	// Docs Controller
	docsGroup := app.Core.Group("/docs")
	docsGroup.Get("/html", docsController.HTML)
}

func (app *App) Run() {
	port := 3000

	var err error
	portString := os.Getenv("PORT")
	if portString != "" {
		port, err = strconv.Atoi(portString)
		if err != nil {
			panic(err)
		}
	}

	if err := app.Core.Listen(fmt.Sprintf(":%v", port)); err != nil {
		panic(err)
	}
	app.client.Disconnect(app.Ctx)
}
