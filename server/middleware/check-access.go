package middleware

import (
	"errors"
	"flappy-bird-server/lib"
	"flappy-bird-server/model"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
)

func CheckAccess(w http.ResponseWriter, r *http.Request) (model.User, error) {
	log.Println("Checking access")
	// cookie, err := r.Cookie("token")
	tokenArr := strings.Split(r.Header.Get("Authorization"), " ")
	if len(tokenArr) < 1 {
		return model.User{}, errors.New("unauthorized")
	}
	tokenString := tokenArr[1]

	if tokenString == "" {
		return model.User{}, errors.New("unauthorized")
	}

	claims := &model.TokenPayload{}

	token, err := jwt.ParseWithClaims(tokenString, claims, (func(t *jwt.Token) (interface{}, error) {
		// return t, nil
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		// Return the secret key used for signing
		return []byte(os.Getenv("SECRET")), nil

	}))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return model.User{}, err
	}

	if !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return model.User{}, errors.New("unauthorized")
	}

	var user model.User
	err = lib.Pool.QueryRow(r.Context(), `SELECT id, name, email, "inrBalance", "solanaBalance" FROM public.users WHERE id = $1`, claims.Id).Scan(&user.Id, &user.Name, &user.Email, &user.INRBalance, &user.SolanaBalance)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return model.User{}, errors.New("internal server error")
	}
	return user, nil
}
