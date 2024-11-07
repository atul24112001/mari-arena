package gametype

import (
	"flappy-bird-server/middleware"

	"github.com/gofiber/fiber/v2"
)

func Router(api fiber.Router) {
	gameTypeRoute := api.Group("/game-type")

	gameTypeRoute.Post("/", middleware.CheckAccess, addGameType)
	gameTypeRoute.Get("/", getGameTypes)
}
