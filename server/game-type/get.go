package gametype

import (
	"flappy-bird-server/lib"
	"flappy-bird-server/model"

	"github.com/gofiber/fiber/v2"
)

func getGameTypes(c *fiber.Ctx) error {
	gameTypes := []model.GameType{}

	rows, err := lib.Pool.Query(c.Context(), "SELECT id, title, entry, winner, currency FROM public.gametypes")
	if err != nil {
		return c.Status(500).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}
	defer rows.Close()

	for rows.Next() {
		var i model.GameType
		if err := rows.Scan(&i.Id, &i.Title, &i.Entry, &i.Winner, &i.Currency); err != nil {
			return c.Status(500).JSON(map[string]interface{}{
				"message": err.Error(),
			})
		}
		gameTypes = append(gameTypes, i)
	}

	return c.Status(200).JSON(map[string]interface{}{
		"data":    gameTypes,
		"message": "success",
	})
}
