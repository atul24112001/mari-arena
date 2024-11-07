package gametype

import (
	"flappy-bird-server/lib"
	"flappy-bird-server/middleware"
	"flappy-bird-server/model"
	"net/http"

	"github.com/google/uuid"
)

type RequestBody struct {
	Title     string `json:"title"`
	Entry     uint   `json:"entry"`
	Winner    uint   `json:"winner"`
	Currency  string `json:"currency"`
	MaxPlayer uint   `json:"maxPlayer"`
}

func addGameType(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.CheckAccess(w, r)
	if err != nil {
		lib.ErrorJson(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	if !user.IsAdmin {
		lib.ErrorJson(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	var body RequestBody
	err = lib.ReadJsonFromBody(r, w, &body)
	if err != nil {
		lib.ErrorJson(w, http.StatusBadRequest, err.Error(), "")
		return
	}

	if body.Currency != "INR" && body.Currency != "SOL" {
		lib.ErrorJson(w, http.StatusUnauthorized, "Invalid currency input", "")
		return
	}

	gameTypeId, err := uuid.NewRandom()
	if err != nil {
		lib.ErrorJson(w, http.StatusInternalServerError, err.Error(), "")
		return
	}

	var gameType model.GameType
	err = lib.Pool.QueryRow(r.Context(), `INSERT INTO public.gametypes (id, title, entry, winner, currency, "maxPlayer") VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, title, entry, winner, currency, "maxPlayer"`, gameTypeId, body.Title, body.Entry, body.Winner, body.Currency, body.MaxPlayer).Scan(&gameType.Id, &gameType.Title, &gameType.Entry, &gameType.Winner, &gameType.Currency, &gameType.MaxPlayer)
	if err != nil {
		lib.ErrorJson(w, http.StatusInternalServerError, err.Error(), "")
		return
	}
	lib.WriteJson(w, http.StatusOK, map[string]interface{}{
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
