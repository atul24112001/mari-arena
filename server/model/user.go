package model

import (
	"github.com/golang-jwt/jwt"
)

type AddUserData struct {
	UserId string `json:"userId"`
}

type AddUserModel struct {
	Type string      `json:"type"`
	Data AddUserData `json:"data"`
}

type User struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	INRBalance    uint   `json:"inrBalance"`
	SolanaBalance uint   `json:"solanaBalance"`
}

type TokenPayload struct {
	Id string `json:"id"`
	jwt.StandardClaims
}
