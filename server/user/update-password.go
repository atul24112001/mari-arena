package user

import (
	"flappy-bird-server/lib"
	"flappy-bird-server/model"

	"github.com/gofiber/fiber/v2"
)

type UpdatePasswordRequestBody struct {
	Password string `json:"password"`
}

func updatePassword(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]interface{}{
			"message": "Unauthorized",
		})
	}

	var body UpdatePasswordRequestBody
	err := c.BodyParser(body)
	if err != nil {
		return c.Status(500).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}

	if len(body.Password) < 8 {
		return c.Status(400).JSON(map[string]interface{}{
			"message": "password length should be more than 7",
		})

	}

	hashedPassword := lib.HashString(user.Email + body.Password)
	_, err = lib.Pool.Exec(c.Context(), "UPDATE public.users SET password = $2 WHERE id = $1", user.Id, hashedPassword)
	if err != nil {
		return c.Status(500).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"message": "Password updated successfully",
	})
}
