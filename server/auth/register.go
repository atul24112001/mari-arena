package auth

import (
	"flappy-bird-server/lib"
	"flappy-bird-server/model"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mr-tron/base58/base58"

	"crypto/ed25519"
)

type AuthenticateRequestBody struct {
	Identifier string  `json:"identifier"`
	Password   string  `json:"password"`
	Signature  []uint8 `json:"signature"`
}

func register(w http.ResponseWriter, r *http.Request) {
	var body AuthenticateRequestBody
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
	if err == nil {
		lib.ErrorJson(w, http.StatusBadRequest, "User already exist", "")
		return
	} else if err != pgx.ErrNoRows {
		lib.ErrorJson(w, http.StatusBadRequest, err.Error(), "")
		return
	}

	publicKey, err := base58.Decode(body.Identifier)
	if err != nil {
		lib.ErrorJson(w, http.StatusBadRequest, err.Error(), "")
		return
	}
	// signature, err := hex.DecodeString(body.Signature)
	// if err != nil {
	// 	lib.ErrorJson(w, http.StatusBadRequest, err.Error(), "")
	// 	return
	// }

	isValid := ed25519.Verify(publicKey, []byte(body.Password), body.Signature)
	if !isValid {
		lib.ErrorJson(w, http.StatusBadRequest, "Invalid signature", "")
		return
	}

	newUserId, err := uuid.NewRandom()
	if err != nil {
		lib.ErrorJson(w, http.StatusBadRequest, err.Error(), "")
		return
	}
	passwordHash = lib.HashString(body.Password)
	err = lib.Pool.QueryRow(r.Context(), "INSERT INTO public.users (id, name, email, password) VALUES ($1, $2, $3, $4) RETURNING id, name, email", newUserId.String(), body.Identifier, body.Identifier, passwordHash).Scan(&user.Id, &user.Name, &user.Email)
	if err != nil {
		lib.ErrorJson(w, http.StatusBadRequest, err.Error(), "")
		return
	}
	token, err := lib.GenerateToken(user.Id)
	if err != nil {
		lib.ErrorJson(w, http.StatusBadRequest, err.Error(), "")
		return
	}
	lib.WriteJson(w, http.StatusOK, map[string]interface{}{
		"message": "Registered successfully",
		"token":   token,
		"data":    user,
	})

	// currentPasswordHash := lib.HashString(body.Password)
	// token, err := lib.GenerateToken(user.Id)
	// if err != nil {
	// 	lib.ErrorJson(w, http.StatusBadRequest, err.Error(), "")
	// 	return
	// }

	// if passwordHash != currentPasswordHash {
	// 	lib.ErrorJson(w, http.StatusBadRequest, "invalid password", "")
	// 	return
	// }

	// gameManager.GetInstance().RedisClient.Set(r.Context(), fmt.Sprintf("mr-balance-%s", user.Id), user.SolanaBalance, 24*time.Hour)

	// lib.WriteJson(w, http.StatusOK, map[string]interface{}{
	// 	"message": "Login successfully",
	// 	"token":   token,
	// 	"data":    user,
	// })
}
