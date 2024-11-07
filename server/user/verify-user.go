package user

import (
	"flappy-bird-server/lib"
	"flappy-bird-server/model"

	"github.com/gofiber/fiber/v2"
)

type RequestBody struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

func verifyUser(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]interface{}{
			"message": "Unauthorized",
		})
	}

	response := map[string]interface{}{
		"message": "success",

		"data": []model.User{user},
	}

	if user.Email == lib.AdminPublicKey {
		response["isAdmin"] = true
	}

	if lib.UnderMaintenance {
		response["underMaintenance"] = true
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
