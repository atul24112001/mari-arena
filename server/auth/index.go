package auth

import (
	"github.com/gorilla/mux"
)

func Handler(r *mux.Router) {
	r.HandleFunc("/register", register).Methods("POST")
	r.HandleFunc("/login", login).Methods("POST")
}
