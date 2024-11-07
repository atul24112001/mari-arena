package gametype

import (
	"flappy-bird-server/lib"
	"flappy-bird-server/model"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type RequestBody struct {
	Title     string `json:"title"`
	Entry     uint   `json:"entry"`
	Winner    uint   `json:"winner"`
	Currency  string `json:"currency"`
	MaxPlayer uint   `json:"maxPlayer"`
}

func addGameType(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]interface{}{
			"message": "Unauthorized",
		})
	}

	if user.Email != lib.AdminPublicKey {
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]interface{}{
			"message": "Unauthorized",
		})
	}

	var body RequestBody
	err := c.BodyParser(body)
	if err != nil {
		return c.Status(400).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}

	if body.Currency != "INR" && body.Currency != "SOL" {
		return c.Status(400).JSON(map[string]interface{}{
			"message": "Invalid currency input",
		})
	}

	gameTypeId, err := uuid.NewRandom()
	if err != nil {
		return c.Status(500).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}

	var gameType model.GameType
	err = lib.Pool.QueryRow(c.Context(), `INSERT INTO public.gametypes (id, title, entry, winner, currency, "maxPlayer") VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, title, entry, winner, currency, "maxPlayer"`, gameTypeId, body.Title, body.Entry, body.Winner, body.Currency, body.MaxPlayer).Scan(&gameType.Id, &gameType.Title, &gameType.Entry, &gameType.Winner, &gameType.Currency, &gameType.MaxPlayer)
	if err != nil {
		return c.Status(500).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}
	return c.Status(200).JSON(map[string]interface{}{
		"message": "Game type create successfully",
		"data": []model.GameType{{
			Id:       gameTypeId.String(),
			Title:    body.Title,
			Entry:    int(body.Entry),
			Winner:   int(body.Winner),
			Currency: body.Currency,
		}},
	})
}
