package main

import (
	"encoding/json"
	"flappy-bird-server/admin"
	"flappy-bird-server/auth"
	gameManager "flappy-bird-server/game-manager"
	gametype "flappy-bird-server/game-type"
	"flappy-bird-server/lib"
	"flappy-bird-server/transaction"
	"flappy-bird-server/user"
	"log"
	"os"
	"runtime"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func handleWebSocket(conn *websocket.Conn) {
	var userId = conn.Params("id")
	log.Println("handler", userId)
	defer func() {
		conn.Close()
	}()
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

		if m["event"] == "ping" {
			user, exist := gameManager.GetInstance().GetUser(userId)
			if exist {
				user.SendMessage("pong", map[string]interface{}{})
			}
		}

		switch messageType {
		case "add-user":
			if !lib.UnderMaintenance {
				gameManager.GetInstance().AddUser(userId, messageData["publicKey"].(string), conn)
			}
		case "join-random-game":
			if !lib.UnderMaintenance {
				gameManager.GetInstance().JoinGame(userId, messageData["gameTypeId"].(string))
			}
		case "update-board":
			gameManager.GetInstance().UpdateBoard(messageData["gameId"].(string), userId)
		case "game-over":
			gameManager.GetInstance().GameOver(messageData["gameId"].(string), userId)
		}
	}
}

func main() {
	runtime.GOMAXPROCS(1)
	lib.ConnectDB()

	app := fiber.New()
	app.Use(logger.New())
	api := app.Group("/api")

	transaction.Router(api)

	api.Use(cors.New(cors.Config{
		AllowOrigins: os.Getenv("FRONTEND_URL"),
	}))

	user.Router(api)
	auth.Router(api)
	gametype.Router(api)
	admin.Router(api)

	app.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			log.Println("Upgrade ws")
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:id", websocket.New(handleWebSocket, websocket.Config{
		Origins: []string{os.Getenv("FRONTEND_URL")},
	}))

	log.Fatal(app.Listen(":8080"))
}

// nodemon --exec go run main.go --signal SIGTERM
