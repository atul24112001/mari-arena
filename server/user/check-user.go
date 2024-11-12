package user

import (
	"flappy-bird-server/lib"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

func CheckUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	if id == "" {
		lib.ErrorJson(w, http.StatusBadRequest, "Id is required", "")
		return
	}

	log.Println(id)
	var userId string
	err := lib.Pool.QueryRow(r.Context(), "SELECT id FROM public.users WHERE email = $1", id).Scan(&userId)

	if err != nil {
		if err == pgx.ErrNoRows {
			log.Println(err.Error())

			lib.ErrorJson(w, http.StatusNotFound, "User not found", "")
			return
		}
		lib.ErrorJson(w, http.StatusBadRequest, "Something went wrong", "")
		return
	}

	lib.WriteJson(w, http.StatusOK, map[string]string{
		"message": "User exist",
	})
}
