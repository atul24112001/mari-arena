package user

import (
	"errors"
	"flappy-bird-server/lib"
	"flappy-bird-server/middleware"
	"net/http"
)

type UpdatePasswordRequestBody struct {
	Password string `json:"password"`
}

func updatePassword(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.CheckAccess(w, r)
	if err != nil {
		lib.ErrorJson(w, err)
		return
	}

	var body UpdatePasswordRequestBody
	err = lib.ReadJsonFromBody(r, w, &body)
	if err != nil {
		lib.ErrorJson(w, err)
		return
	}

	if len(body.Password) < 8 {
		lib.ErrorJson(w, errors.New("password length should be more than 7"))
		return
	}

	hashedPassword := lib.HashString(user.Email + body.Password)
	_, err = lib.Pool.Exec(r.Context(), "UPDATE public.users SET password = $2 WHERE id = $1", user.Id, hashedPassword)
	if err != nil {
		lib.ErrorJson(w, err)
		return
	}
	lib.WriteJson(w, http.StatusOK, map[string]interface{}{
		"message": "Password updated successfully",
	})
}
