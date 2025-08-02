package routes

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
	"roflbeacon2/app/api"
)

func NotFoundRoute(a *fiber.App) {
	a.Use(
		func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusNotFound).JSON(api.General{
				Error:      true,
				Msg:        "route not found",
				StatusCode: http.StatusNotFound,
			})
		},
	)
}
