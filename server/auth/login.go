package auth

import (
	gameManager "flappy-bird-server/game-manager"
	"flappy-bird-server/lib"
	"flappy-bird-server/model"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

type LoginRequestBody struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func login(w http.ResponseWriter, r *http.Request) {
	log.Println("Login")
	var body LoginRequestBody
	err := lib.ReadJsonFromBody(r, w, &body)
	if err != nil {
		lib.ErrorJson(w, http.StatusBadRequest, err.Error(), "")
		return
	}
	if len(body.Password) < 7 {
		lib.ErrorJson(w, http.StatusBadRequest, "password length should be more than 7", "")
		return
	}
	if len(body.Password) > 15 {
		lib.ErrorJson(w, http.StatusBadRequest, "password length should be less then 16", "")
		return
	}

	var user model.User
	var passwordHash string

	getUserDetailsQuery := `SELECT id, name, email, "inrBalance", "solanaBalance", password  FROM public.users WHERE email = $1`
	err = lib.Pool.QueryRow(r.Context(), getUserDetailsQuery, body.Identifier).Scan(&user.Id, &user.Name, &user.Email, &user.INRBalance, &user.SolanaBalance, &passwordHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			lib.ErrorJson(w, http.StatusBadRequest, "User not found with this public key", "")
			return
		}
		lib.ErrorJson(w, http.StatusBadRequest, err.Error(), "")
		return
	}
	currentPasswordHash := lib.HashString(body.Password)
	token, err := lib.GenerateToken(user.Id)
	if err != nil {
		lib.ErrorJson(w, http.StatusBadRequest, err.Error(), "")
		return
	}

	if passwordHash != currentPasswordHash {
		lib.ErrorJson(w, http.StatusBadRequest, "invalid password", "")
		return
	}

	gameManager.GetInstance().RedisClient.Set(r.Context(), fmt.Sprintf("mr-balance-%s", user.Id), user.SolanaBalance, 24*time.Hour)

	lib.WriteJson(w, http.StatusOK, map[string]interface{}{
		"message": "Login successfully",
		"token":   token,
		"data":    user,
	})
}
