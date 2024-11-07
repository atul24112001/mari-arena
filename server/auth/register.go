package auth

import (
	"errors"
	"flappy-bird-server/lib"
	"flappy-bird-server/model"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AuthenticateRequestBody struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func authenticate(c *fiber.Ctx) error {
	var body AuthenticateRequestBody
	err := c.BodyParser(body)
	if err != nil {
		return c.Status(400).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}

	if len(body.Password) < 7 {
		return c.Status(400).JSON(map[string]interface{}{
			"message": "password length should be more than 7",
		})
	}
	if len(body.Password) > 15 {
		return c.Status(400).JSON(map[string]interface{}{
			"message": "password length should be less then 16",
		})
	}

	var user model.User
	var passwordHash string

	getUserDetailsQuery := `SELECT id, name, email, "inrBalance", "solanaBalance", password  FROM public.users WHERE email = $1`
	err = lib.Pool.QueryRow(c.Context(), getUserDetailsQuery, body.Identifier).Scan(&user.Id, &user.Name, &user.Email, &user.INRBalance, &user.SolanaBalance, &passwordHash)
	if err != nil {
		if err == pgx.ErrNoRows {
			newUserId, err := uuid.NewRandom()
			if err != nil {
				return c.Status(500).JSON(map[string]interface{}{
					"message": err.Error(),
				})
			}
			passwordHash = lib.HashString(body.Password)
			err = lib.Pool.QueryRow(c.Context(), "INSERT INTO public.users (id, name, email, password) VALUES ($1, $2, $3, $4) RETURNING id, name, email", newUserId.String(), body.Identifier, body.Identifier, passwordHash).Scan(&user.Id, &user.Name, &user.Email)
			if err != nil {
				return c.Status(500).JSON(map[string]interface{}{
					"message": err.Error(),
				})
			}
			token, err := lib.GenerateToken(user.Id)
			if err != nil {
				return c.Status(500).JSON(map[string]interface{}{
					"message": err.Error(),
				})
			}
			return c.Status(200).JSON(map[string]interface{}{
				"message": "Registered successfully",
				"token":   token,
				"data":    user,
			})
		} else {
			return c.Status(500).JSON(map[string]interface{}{
				"message": err.Error(),
			})
		}
	}
	currentPasswordHash := lib.HashString(body.Password)
	token, err := lib.GenerateToken(user.Id)
	if err != nil {
		return c.Status(500).JSON(map[string]interface{}{
			"message": err.Error(),
		})
	}

	if passwordHash != currentPasswordHash {
		return c.Status(400).JSON(map[string]interface{}{
			"message": errors.New("invalid password"),
		})
	}

	return c.Status(200).JSON(map[string]interface{}{
		"message": "Login successfully",
		"token":   token,
		"data":    user,
	})
}
