package auth

import (
	"flappy-bird-server/lib"
	"net/http"
)

type LoginRequestBody struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func login(w http.ResponseWriter, r *http.Request) {
	var body LoginRequestBody
	err = lib.ReadJsonFromBody(r, w, &body)
	if err != nil {
		lib.ErrorJson(w, err)
		return
	}
}
