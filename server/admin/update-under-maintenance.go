package admin

import (
	"flappy-bird-server/lib"

	"github.com/gofiber/fiber/v2"
)

func updateUnderMaintenance(c *fiber.Ctx) error {
	lib.UnderMaintenance = !lib.UnderMaintenance
	return c.JSON(map[string]interface{}{
		"message":       "Maintenance status updated successfully",
		"currentStatus": lib.UnderMaintenance,
	})
}
