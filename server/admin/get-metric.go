package admin

import (
	gameManager "flappy-bird-server/game-manager"
	"flappy-bird-server/lib"
	"flappy-bird-server/middleware"
	"net/http"
)

type Game struct {
	Id     string   `json:"id"`
	Status string   `json:"status"`
	Users  []string `json:"users"`
}

func GetMetrics(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.CheckAccess(w, r)
	if err != nil {
		lib.ErrorJson(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	if !user.IsAdmin {
		lib.ErrorJson(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

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

	lib.WriteJson(w, http.StatusOK, map[string]interface{}{
		"message": "success",
		"data": map[string]interface{}{
			"ongoingGames": ongoingGames,
			"activeUsers":  activeUsers,
		},
	})
}
