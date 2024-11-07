package gametype

import (
	"flappy-bird-server/lib"
	"flappy-bird-server/model"
	"net/http"
)

func getGameTypes(w http.ResponseWriter, r *http.Request) {
	gameTypes := []model.GameType{}
	rows, err := lib.Pool.Query(r.Context(), "SELECT id, title, entry, winner, currency FROM public.gametypes")
	if err != nil {
		lib.ErrorJson(w, http.StatusInternalServerError, err.Error(), "")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var i model.GameType
		if err := rows.Scan(&i.Id, &i.Title, &i.Entry, &i.Winner, &i.Currency); err != nil {
			lib.ErrorJson(w, http.StatusInternalServerError, err.Error(), "")
			return
		}
		gameTypes = append(gameTypes, i)
	}
	lib.WriteJson(w, http.StatusOK, map[string]interface{}{
		"data":    gameTypes,
		"message": "success",
	})

}
