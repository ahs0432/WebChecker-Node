package routes

import (
	"false.kr/WebChecker-Node/controllers"
	"github.com/gofiber/fiber/v2"
)

func Router() *fiber.App {
	app := fiber.New(fiber.Config{
		Prefork:       true,
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Chance",
		AppName:       "WebChecker-Node",
	})

	app.Get("/", func(c *fiber.Ctx) error {
		err := c.SendString("API Health OK!!")
		return err
	})

	app.Post("/api", controllers.API)
	app.Get("/images/:targetId", controllers.GetImage)
	return app
}
