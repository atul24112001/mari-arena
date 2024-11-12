package gametype

import (
	"flappy-bird-server/lib"
	"net/http"
)

// func Handler(r *mux.Router) {
// 	r.HandleFunc("/", getGameTypes).Methods("GET")
// 	r.HandleFunc("/", addGameType).Methods("POST")
// }

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		addGameType(w, r)
		return
	} else if r.Method == http.MethodGet {
		getGameTypes(w, r)
		return
	}
	lib.ErrorJson(w, 405, "Method not allowed", "")
}
