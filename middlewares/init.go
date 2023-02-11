package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"treehole_backend/config"
)

func RegisterMiddlewares(app *fiber.App) {
	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	if config.Config.Mode != "bench" {
		app.Use(logger.New())
	}
	app.Use(cors.New(cors.Config{AllowOrigins: "*"}))
}
