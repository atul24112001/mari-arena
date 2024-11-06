package auth

import (
	"errors"
	"flappy-bird-server/lib"
	"flappy-bird-server/model"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AuthenticateRequestBody struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func authenticate(w http.ResponseWriter, r *http.Request) {
	var body AuthenticateRequestBody
	err := lib.ReadJsonFromBody(r, w, &body)
	if err != nil {
		lib.ErrorJson(w, err)
		return
	}

	if len(body.Password) < 7 {
		lib.ErrorJsonWithCode(w, errors.New("password length should be more than 7"))
		return
	}
	if len(body.Password) > 15 {
		lib.ErrorJsonWithCode(w, errors.New("password length should be less then 16"))
		return
	}

	var user model.User
	var passwordHash string

	getUserDetailsQuery := `SELECT id, name, email, "inrBalance", "solanaBalance", password  FROM public.users WHERE email = $1`
	err = lib.Pool.QueryRow(r.Context(), getUserDetailsQuery, body.Identifier).Scan(&user.Id, &user.Name, &user.Email, &user.INRBalance, &user.SolanaBalance, &passwordHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			newUserId, err := uuid.NewRandom()
			if err != nil {
				lib.ErrorJsonWithCode(w, err, 500)
				return
			}
			passwordHash = lib.HashString(body.Password)
			err = lib.Pool.QueryRow(r.Context(), "INSERT INTO public.users (id, name, email, password) VALUES ($1, $2, $3, $4) RETURNING id, name, email", newUserId.String(), body.Identifier, body.Identifier, passwordHash).Scan(&user.Id, &user.Name, &user.Email)
			if err != nil {
				lib.ErrorJsonWithCode(w, err, 500)
				return
			}
			token, err := lib.GenerateToken(user.Id)
			if err != nil {
				lib.ErrorJsonWithCode(w, err, 500)
				return
			}
			lib.WriteJson(w, http.StatusOK, map[string]interface{}{
				"message": "Registered successfully",
				"token":   token,
				"data":    user,
			})
			return
		} else {

			lib.ErrorJsonWithCode(w, err, 500)
			return
		}
	}
	currentPasswordHash := lib.HashString(body.Password)
	token, err := lib.GenerateToken(user.Id)
	if err != nil {
		lib.ErrorJsonWithCode(w, err, 500)
		return
	}

	if passwordHash != currentPasswordHash {
		lib.ErrorJsonWithCode(w, errors.New("invalid password"), 400)
		return
	}

	lib.WriteJson(w, http.StatusOK, map[string]interface{}{
		"message": "Login successfully",
		"token":   token,
		"data":    user,
	})

}
