package admin

import (
	"github.com/gorilla/mux"
)

func Handler(r *mux.Router) {
	r.HandleFunc("/metric", GetMetrics).Methods("GET")
	r.HandleFunc("/maintenance", UpdateUnderMaintenance).Methods("GET")
}
