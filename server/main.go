package main

import (
	"encoding/json"
	"flappy-bird-server/auth"
	gameManager "flappy-bird-server/game-manager"
	gametype "flappy-bird-server/game-type"
	"flappy-bird-server/lib"
	"flappy-bird-server/transaction"
	"flappy-bird-server/user"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func handleWebSocket(conn *websocket.Conn) {
	log.Println()
	var userId = conn.Params("id")
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			gameManager.GetInstance().DeleteUser(conn, userId)
			break
		}

		m := make(map[string]interface{})
		err = json.Unmarshal(message, &m)
		if err != nil {
			log.Println("Error message:", err.Error())
			return
		}

		messageType := m["type"]
		messageData, ok := m["data"].(map[string]interface{})

		if !ok {
			log.Println("Something went wrong while parsing message data.")
			return
		}

		switch messageType {
		case "add-user":
			gameManager.GetInstance().AddUser(userId, messageData["publicKey"].(string), conn)
		case "join-random-game":
			gameManager.GetInstance().JoinGame(userId, messageData["gameTypeId"].(string))
		case "update-board":
			gameManager.GetInstance().UpdateBoard(messageData["gameId"].(string), userId)
		case "game-over":
			gameManager.GetInstance().GameOver(messageData["gameId"].(string), userId)
		}
	}
}

func httpCorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONTEND_URL"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	runtime.GOMAXPROCS(1)
	lib.ConnectDB()

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		origin, ok := c.Locals("origin").(string)
		if origin != os.Getenv("FRONTEND_URL") || !ok {
			return c.Status(403).JSON(map[string]interface{}{
				"message": "Cors error",
			})
		}
		return c.Next()
	})

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:id", websocket.New(handleWebSocket))

	api := app.Group("/api")
	transaction.Router(api)

	user.Router(api)
	auth.Router(api)
	gametype.Router(api)

	log.Fatal(app.Listen(":8080"))
}

// nodemon --exec go run main.go --signal SIGTERM
