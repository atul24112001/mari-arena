package admin

import (
	"flappy-bird-server/middleware"

	"github.com/gofiber/fiber/v2"
)

func Router(api fiber.Router) {
	adminRoute := api.Group("/admin")

	adminRoute.Get("/metric", middleware.CheckAccess, middleware.CheckIsAdmin, getMetrics)
	adminRoute.Get("/maintenance", middleware.CheckAccess, middleware.CheckIsAdmin, updateUnderMaintenance)
}
