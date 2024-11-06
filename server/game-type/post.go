package gametype

import (
	"errors"
	"flappy-bird-server/lib"
	"flappy-bird-server/middleware"
	"flappy-bird-server/model"
	"log"
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
	log.Println("23")
	user, err := middleware.CheckAccess(w, r)
	if err != nil {
		log.Println("25")
		lib.ErrorJsonWithCode(w, err, http.StatusUnauthorized)
		return
	}

	if user.Email != lib.AdminPublicKey {
		log.Println("31")
		lib.ErrorJsonWithCode(w, errors.New("Unauthorized"), http.StatusUnauthorized)
		return
	}

	var body RequestBody
	if err := lib.ReadJsonFromBody(r, w, &body); err != nil {
		log.Println("38")
		log.Println(err.Error())
		lib.ErrorJsonWithCode(w, err, http.StatusInternalServerError)
		return
	}

	if body.Currency != "INR" && body.Currency != "SOL" {
		log.Println("45")
		lib.ErrorJsonWithCode(w, errors.New("Invalid currency input"), http.StatusBadRequest)
		return
	}

	gameTypeId, err := uuid.NewRandom()
	if err != nil {
		log.Println("52")
		lib.ErrorJsonWithCode(w, errors.New("Something went wrong while generating uuid"), http.StatusInternalServerError)
		return
	}

	var gameType model.GameType
	err = lib.Pool.QueryRow(r.Context(), `INSERT INTO public.gametypes (id, title, entry, winner, currency, "maxPlayer") VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, title, entry, winner, currency, "maxPlayer"`, gameTypeId, body.Title, body.Entry, body.Winner, body.Currency, body.MaxPlayer).Scan(&gameType.Id, &gameType.Title, &gameType.Entry, &gameType.Winner, &gameType.Currency, &gameType.MaxPlayer)
	if err != nil {
		log.Println("60")
		lib.ErrorJsonWithCode(w, err, http.StatusInternalServerError)
		return
	}

	log.Println("65")
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
