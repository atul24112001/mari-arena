package admin

import (
	"flappy-bird-server/lib"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		GetMetrics(w, r)
		return
	}
	if r.Method == http.MethodPost {
		UpdateUnderMaintenance(w, r)
		return
	}
	lib.ErrorJson(w, 405, "Method not allowed", "")
}
