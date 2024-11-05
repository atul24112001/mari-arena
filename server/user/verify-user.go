package user

import (
	"flappy-bird-server/lib"
	"flappy-bird-server/middleware"
	"flappy-bird-server/model"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RequestBody struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

func verifyUser(w http.ResponseWriter, r *http.Request) {
	var body RequestBody
	// var JWT_SECRET = []byte(os.Getenv("SECRET"))

	err := lib.ReadJsonFromBody(r, w, &body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	var user model.User
	getUserDetailsQuery := `SELECT id, name, email, "inrBalance", "solanaBalance"  FROM public.users WHERE email = $1`
	err = lib.Pool.QueryRow(r.Context(), getUserDetailsQuery, body.Identifier).Scan(&user.Id, &user.Name, &user.Email, &user.INRBalance, &user.SolanaBalance)
	if err != nil {
		if err == pgx.ErrNoRows {
			newUserId, err := uuid.NewRandom()
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			err = lib.Pool.QueryRow(r.Context(), "INSERT INTO public.users (id, name, email) VALUES ($1, $2, $3) RETURNING id, name, email", newUserId.String(), body.Name, body.Identifier).Scan(&user.Id, &user.Name, &user.Email)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		} else {
			log.Println(err.Error())
			http.Error(w, err.Error(), 500)
			return
		}
	}

	// expirationTime := time.Now().Add(24 * 30 * time.Hour)

	// claims := &model.TokenPayload{
	// 	Id: user.Id,
	// 	StandardClaims: jwt.StandardClaims{
	// 		ExpiresAt: expirationTime.Unix(),
	// 	},
	// }

	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// tokenString, err := token.SignedString(JWT_SECRET)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// http.SetCookie(w, &http.Cookie{
	// 	Name:     "token",
	// 	Value:    tokenString,
	// 	Expires:  expirationTime,
	// 	Secure:   false, // Use true for production (only HTTPS)
	// 	Path:     "/",
	// 	Domain:   "localhost", // No port should be specified
	// 	HttpOnly: true,
	// 	SameSite: http.SameSiteLaxMode,
	// })
	response := map[string]interface{}{
		"message": "success",
		"data":    []model.User{user},
		// "token":   tokenString,
	}

	if user.Email == lib.AdminPublicKey {
		response["isAdmin"] = true
	}

	lib.WriteJson(w, 200, response)
}

func verifyUser2(w http.ResponseWriter, r *http.Request) {
	user, err := middleware.CheckAccess(w, r)
	if err != nil {
		lib.ErrorJsonWithCode(w, err, 401)
		return
	}

	// expirationTime := time.Now().Add(60 * 24 * 30 * time.Minute)

	// claims := &model.TokenPayload{
	// 	Id: user.Id,
	// 	StandardClaims: jwt.StandardClaims{
	// 		ExpiresAt: expirationTime.Unix(),
	// 	},
	// }

	// token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// var JWT_SECRET = []byte(os.Getenv("SECRET"))
	// tokenString, err := token.SignedString(JWT_SECRET)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	response := map[string]interface{}{
		"message": "success",
		"data":    []model.User{user},
		// "token":   tokenString,
	}

	if user.Email == lib.AdminPublicKey {
		response["isAdmin"] = true
	}

	lib.WriteJson(w, 200, response)
}
