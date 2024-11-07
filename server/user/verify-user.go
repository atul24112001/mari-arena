package user

import (
	"flappy-bird-server/lib"
	"flappy-bird-server/model"

	"github.com/gofiber/fiber/v2"
)

type RequestBody struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

// func verifyUser(c *fiber.Ctx) error {
// 	var body RequestBody

// 	if err := c.BodyParser(body); err != nil {
// 		return err
// 	}

// 	var user model.User
// 	getUserDetailsQuery := `SELECT id, name, email, "inrBalance", "solanaBalance"  FROM public.users WHERE email = $1`
// 	err := lib.Pool.QueryRow(c.Context(), getUserDetailsQuery, body.Identifier).Scan(&user.Id, &user.Name, &user.Email, &user.INRBalance, &user.SolanaBalance)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			newUserId, err := uuid.NewRandom()
// 			if err != nil {
// 				return c.Status(500).JSON(map[string]interface{}{
// 					"message": err.Error(),
// 				})
// 			}
// 			err = lib.Pool.QueryRow(c.Context(), "INSERT INTO public.users (id, name, email) VALUES ($1, $2, $3) RETURNING id, name, email", newUserId.String(), body.Name, body.Identifier).Scan(&user.Id, &user.Name, &user.Email)
// 			if err != nil {
// 				return c.Status(500).JSON(map[string]interface{}{
// 					"message": err.Error(),
// 				})
// 			}
// 		} else {
// 			log.Println(err.Error())
// 			return c.Status(500).JSON(map[string]interface{}{
// 				"message": err.Error(),
// 			})
// 		}
// 	}

// 	// expirationTime := time.Now().Add(24 * 30 * time.Hour)

// 	// claims := &model.TokenPayload{
// 	// 	Id: user.Id,
// 	// 	StandardClaims: jwt.StandardClaims{
// 	// 		ExpiresAt: expirationTime.Unix(),
// 	// 	},
// 	// }

// 	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	// tokenString, err := token.SignedString(JWT_SECRET)
// 	// if err != nil {
// 	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
// 	// 	return
// 	// }

// 	// http.SetCookie(w, &http.Cookie{
// 	// 	Name:     "token",
// 	// 	Value:    tokenString,
// 	// 	Expires:  expirationTime,
// 	// 	Secure:   false, // Use true for production (only HTTPS)
// 	// 	Path:     "/",
// 	// 	Domain:   "localhost", // No port should be specified
// 	// 	HttpOnly: true,
// 	// 	SameSite: http.SameSiteLaxMode,
// 	// })
// 	response := map[string]interface{}{
// 		"message": "success",
// 		"data":    []model.User{user},
// 		// "token":   tokenString,
// 	}

// 	if user.Email == lib.AdminPublicKey {
// 		response["isAdmin"] = true
// 	}

// 	// return
// 	// lib.WriteJson(w, 200, response)
// }

func verifyUser(c *fiber.Ctx) error {
	user, ok := c.Locals("user").(model.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(map[string]interface{}{
			"message": "Unauthorized",
		})
	}

	response := map[string]interface{}{
		"message": "success",
		"data":    []model.User{user},
	}

	if user.Email == lib.AdminPublicKey {
		response["isAdmin"] = true
	}

	return c.Status(fiber.StatusUnauthorized).JSON(response)
}

// func verifyUser3(c *fiber.Ctx) {
// 	user, err := middleware.CheckAccess(w, r)
// 	if err != nil {
// 		lib.ErrorJsonWithCode(w, err, 401)
// 		return
// 	}

// 	response := map[string]interface{}{
// 		"message": "success",
// 		"data":    []model.User{user},
// 		// "token":   tokenString,
// 	}

// 	if user.Email == lib.AdminPublicKey {
// 		response["isAdmin"] = true
// 	}

// 	lib.WriteJson(w, 200, response)
// }
