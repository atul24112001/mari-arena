package auth

import (
	"flappy-bird-server/lib"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		authenticate(w, r)
		return
	}
	lib.ErrorJson(w, 405, "Method not allowed", "")
}
