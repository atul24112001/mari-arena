package user

import (
	"github.com/gorilla/mux"
)

func Handler(r *mux.Router) {
	r.HandleFunc("/me", verifyUser).Methods("GET")
	r.HandleFunc("/{id}", CheckUser).Methods("GET")
}

// func Handler() {
// 	log.Println("User route")
// 	w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONTEND_URL"))
// 	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
// 	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

// 	if r.Method == http.MethodOptions {
// 		w.WriteHeader(http.StatusNoContent)
// 		return
// 	}
// 	if r.Method == http.MethodPost {
// 		verifyUser(w, r)
// 		return
// 	}
// 	if r.Method == http.MethodGet {
// 		checkUser(w, r)
// 		return
// 	}
// 	lib.ErrorJson(w, 405, "Method not allowed", "")
// }
