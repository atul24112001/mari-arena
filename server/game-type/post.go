package gametype

import (
	"flappy-bird-server/lib"
	"flappy-bird-server/middleware"
	"net/http"
)

type RequestBody struct {
}

func addGameType(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.CheckAccess(w, r)
	if err != nil {
		lib.ErrorJsonWithCode(w, err, http.StatusBadRequest)
		return
	}
	var body RequestBody
	if err := lib.ReadJsonFromBody(r, w, &body); err != nil {
		lib.ErrorJsonWithCode(w, err, http.StatusBadRequest)
		return
	}
}
