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

type User struct {
	IsAdmin       bool   `json:"isAdmin"`
	Id            string `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	INRBalance    uint   `json:"inrBalance"`
	SolanaBalance uint   `json:"solanaBalance"`
}

func CheckAccess(w http.ResponseWriter, r *http.Request) (User, error) {
	tokenArr := strings.Split(r.Header.Get("Authorization"), " ")
	if len(tokenArr) < 2 {
		return User{}, errors.New("unauthorized")
	}
	tokenString := tokenArr[1]

	if tokenString == "" {
		return User{}, errors.New("unauthorized")
	}

	claims := &model.TokenPayload{}

	token, err := jwt.ParseWithClaims(tokenString, claims, (func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(os.Getenv("SECRET")), nil

	}))
	if err != nil {
		return User{}, err
	}
	if !token.Valid {
		return User{}, errors.New("unauthorized")
	}
	var user User
	err = lib.Pool.QueryRow(r.Context(), `SELECT id, name, email, "inrBalance", "solanaBalance" FROM public.users WHERE id = $1`, claims.Id).Scan(&user.Id, &user.Name, &user.Email, &user.INRBalance, &user.SolanaBalance)
	if err != nil {
		log.Println(err.Error())
		return User{}, errors.New("internal server error")
	}
	if user.Email == lib.AdminPublicKey {
		user.IsAdmin = true
	}
	return user, nil
}
