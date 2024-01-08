package publicapi

import "github.com/gofiber/fiber/v2"

func ServiceRegistration() func(app *fiber.App) {
	return func(app *fiber.App) {
		debugGroup := app.Group("/api/debug")
		debugGroup.Get("/streams", GETDebugListStreams)
		
		app.Get("/healthcheck", GETHealthcheck)
	}
}
