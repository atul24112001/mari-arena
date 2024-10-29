package user

import "net/http"

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		verifyUser(w, r)
		return
	}
	http.Error(w, "method not allowed", 400)
}
