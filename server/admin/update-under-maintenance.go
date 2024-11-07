package admin

import (
	"flappy-bird-server/lib"
	"flappy-bird-server/middleware"
	"net/http"
)

func UpdateUnderMaintenance(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.CheckAccess(w, r)
	if err != nil {
		lib.ErrorJson(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	if !user.IsAdmin {
		lib.ErrorJson(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	lib.UnderMaintenance = !lib.UnderMaintenance
	lib.WriteJson(w, http.StatusOK, map[string]interface{}{
		"message":       "Maintenance status updated successfully",
		"currentStatus": lib.UnderMaintenance,
	})
}
