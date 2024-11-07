package user

import (
	"flappy-bird-server/lib"
	"net/http"
)

// func Router(api fiber.Router) {
// 	userRoute := api.Group("/user")

// 	// userRoute.Post("/", middleware.CheckAccess, verifyUser)
// 	// userRoute.Post("/", middleware.CheckAccess, updatePassword)
// }

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		verifyUser(w, r)
		return
	}
	lib.ErrorJson(w, 405, "Method not allowed", "")
}
