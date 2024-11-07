package transaction

import (
	"flappy-bird-server/lib"
	"net/http"
)

//	func Router(api fiber.Router) {
//		gameTypeRoute := api.Group("/transaction")
//		gameTypeRoute.Post("/", verifyTransaction)
//	}
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		verifyTransaction(w, r)
		return
	}
	lib.ErrorJson(w, 405, "Method not allowed", "")
}
