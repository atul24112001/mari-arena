package gametype

import (
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		addGameType(w, r)
		return
	}
	if r.Method == http.MethodGet {
		getGameTypes(w, r)
		return
	}
	http.Error(w, "method not allowed", 400)
}
