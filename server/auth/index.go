package auth

import (
	"github.com/gofiber/fiber/v2"
)

func Router(api fiber.Router) {
	authRoute := api.Group("/auth")

	authRoute.Post("/", authenticate)
}
