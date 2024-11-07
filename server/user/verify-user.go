package user

import (
	"flappy-bird-server/lib"
	"flappy-bird-server/middleware"
	"net/http"
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

	if user.Email == lib.AdminPublicKey {
		response["isAdmin"] = true
	}

	if lib.UnderMaintenance {
		response["underMaintenance"] = true
	}
	lib.WriteJson(w, http.StatusOK, response)
}
