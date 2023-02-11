package apis

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func RegisterRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/api")
	})
	// docs
	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/index.html")
	})
	app.Get("/docs/*", swagger.HandlerDefault)

	// meta
	routes := app.Group("/api")
	routes.Get("/", Index)

	// token
	routes.Post("/login", Login)
	routes.Get("/logout", Logout)
	routes.Post("/refresh", Refresh)

	// account management
	routes.Get("/verify/email", VerifyWithEmail)
	routes.Post("/register", Register)
	routes.Put("/register", ChangePassword)

	// user info
	routes.Get("/users/me", GetCurrentUser)

	// hole api
	routes.Get("/holes", ListHoles)
	routes.Post("/holes", CreateHole)
	routes.Delete("/holes/:id", DeleteHole)

	// floor api
	routes.Get("/holes/:id/floors", ListFloorsInAHole)
	routes.Get("/floors/:id", GetFloor)
	routes.Post("/holes/:id/floors", CreateFloor)
	routes.Put("/holes/:id", ModifyFloor)
	routes.Delete("/holes/:id", DeleteFloor)
}
