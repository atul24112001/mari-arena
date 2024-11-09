package user

import (
	gameManager "flappy-bird-server/game-manager"
	"flappy-bird-server/lib"
	"flappy-bird-server/middleware"
	"fmt"
	"net/http"
	"time"
)

type RequestBody struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

func verifyUser(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.CheckAccess(w, r)
	if err != nil {
		lib.ErrorJson(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	response := map[string]interface{}{
		"message": "success",
		"data":    []middleware.User{user},
	}

	gameManager.GetInstance().RedisClient.Set(r.Context(), fmt.Sprintf("mr-balance-%s", user.Id), user.SolanaBalance, 24*time.Hour)

	if user.Email == lib.AdminPublicKey {
		response["isAdmin"] = true
	}

	if lib.UnderMaintenance {
		response["underMaintenance"] = true
	}
	lib.WriteJson(w, http.StatusOK, response)
}
