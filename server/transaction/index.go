package transaction

import (
	"github.com/gofiber/fiber/v2"
)

func Router(api fiber.Router) {
	gameTypeRoute := api.Group("/transaction")
	gameTypeRoute.Post("/", verifyTransaction)
}
