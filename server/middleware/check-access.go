package middleware

import (
	"flappy-bird-server/lib"
	"flappy-bird-server/model"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

func CheckAccess(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).SendString("Missing or invalid authorization header")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &model.TokenPayload{}

	token, err := jwt.ParseWithClaims(tokenString, claims, (func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(os.Getenv("SECRET")), nil

	}))
	if err != nil {
		return c.Status(500).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}

	if !token.Valid {
		return c.Status(401).JSON(map[string]interface{}{
			"message": "Unauthorized",
		})
	}

	var user model.User
	err = lib.Pool.QueryRow(c.Context(), `SELECT id, name, email, "inrBalance", "solanaBalance" FROM public.users WHERE id = $1`, claims.Id).Scan(&user.Id, &user.Name, &user.Email, &user.INRBalance, &user.SolanaBalance)
	if err != nil {
		return c.Status(500).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}

	c.Locals("user", user)
	return c.Next()
}

// func CheckAccess(w http.ResponseWriter, r *http.Request) (model.User, error) {
// 	tokenArr := strings.Split(r.Header.Get("Authorization"), " ")
// 	if len(tokenArr) < 2 {
// 		return model.User{}, errors.New("unauthorized")
// 	}
// 	tokenString := tokenArr[1]

// 	if tokenString == "" {
// 		return model.User{}, errors.New("unauthorized")
// 	}

// claims := &model.TokenPayload{}

// token, err := jwt.ParseWithClaims(tokenString, claims, (func(t *jwt.Token) (interface{}, error) {
// 	// return t, nil
// 	if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
// 		return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
// 	}
// 	// Return the secret key used for signing
// 	return []byte(os.Getenv("SECRET")), nil

// }))
// if err != nil {
// 	w.WriteHeader(http.StatusInternalServerError)
// 	return model.User{}, err
// }

// if !token.Valid {
// 	w.WriteHeader(http.StatusUnauthorized)
// 	return model.User{}, errors.New("unauthorized")
// }

// var user model.User
// err = lib.Pool.QueryRow(r.Context(), `SELECT id, name, email, "inrBalance", "solanaBalance" FROM public.users WHERE id = $1`, claims.Id).Scan(&user.Id, &user.Name, &user.Email, &user.INRBalance, &user.SolanaBalance)
// if err != nil {
// 	w.WriteHeader(http.StatusInternalServerError)
// 	return model.User{}, errors.New("internal server error")
// }
// 	return user, nil
// }
