package user

import (
	"flappy-bird-server/middleware"

	"github.com/gofiber/fiber/v2"
)

func Router(api fiber.Router) {
	userRoute := api.Group("/user")

	userRoute.Post("/", middleware.CheckAccess, verifyUser)
	userRoute.Post("/", middleware.CheckAccess, updatePassword)
}
