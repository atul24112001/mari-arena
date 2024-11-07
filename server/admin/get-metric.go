package admin

import (
	gameManager "flappy-bird-server/game-manager"

	"github.com/gofiber/fiber/v2"
)

type Game struct {
	Id     string   `json:"id"`
	Status string   `json:"status"`
	Users  []string `json:"users"`
}

func getMetrics(c *fiber.Ctx) error {
	ongoingGames := []Game{}
	for gameId, game := range gameManager.GetInstance().StartedGames {
		users := make([]string, 0, len(game.Users))
		for k := range game.Users {
			users = append(users, k)
		}

		ongoingGames = append(ongoingGames, Game{
			Id:     gameId,
			Status: game.Status,
			Users:  users,
		})
	}

	activeUsers := make([]string, 0, len(gameManager.GetInstance().Users))
	for k := range gameManager.GetInstance().Users {
		activeUsers = append(activeUsers, k)
	}

	return c.Status(200).JSON(map[string]interface{}{
		"message": "success",
		"data": map[string]interface{}{
			"ongoingGames": ongoingGames,
			"activeUsers":  activeUsers,
		},
	})
}
