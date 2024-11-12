package main

import (
	"context"
	"encoding/json"
	"flappy-bird-server/admin"
	"flappy-bird-server/auth"
	gameManager "flappy-bird-server/game-manager"
	gametype "flappy-bird-server/game-type"
	"flappy-bird-server/lib"
	"flappy-bird-server/transaction"
	"flappy-bird-server/user"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("origin")
		return origin == os.Getenv("FRONTEND_URL")
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade:", err)
		return
	}
	defer conn.Close()

	for {
		n, message, err := conn.ReadMessage()
		if err != nil {
			targetUserId, exist := gameManager.GetInstance().UserConnectionMap[conn]
			if exist {
				delete(gameManager.GetInstance().UserConnectionMap, conn)
				gameManager.GetInstance().GameQueue.Enqueue(context.Background(), map[string]interface{}{
					"type": "delete-user",
					"data": map[string]string{
						"userId": targetUserId,
					},
				})
			}
			break
		}

		log.Println(n)
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
		log.Println("messageType", messageType)
		switch messageType {
		case "add-user":
			gameManager.GetInstance().AddUser(messageData["userId"].(string), messageData["publicKey"].(string), conn)
		case "join-random-game":
			if lib.UnderMaintenance {
				targetUser, exist := gameManager.GetInstance().GetUser(messageData["userId"].(string))
				if exist {
					targetUser.SendMessage("error", map[string]interface{}{
						"message": "We are under maintenance please try after some time",
					})
				}
			} else {
				log.Println("joining game", messageData["userId"].(string))
				gameManager.GetInstance().GameQueue.Enqueue(context.Background(), map[string]interface{}{
					"type": "join-game",
					"data": messageData,
				})
			}
		case "update-board":
			gameManager.GetInstance().UpdateBoard(messageData["gameId"].(string), messageData["userId"].(string))
		case "game-over":
			gameManager.GetInstance().GameOver(messageData["gameId"].(string), messageData["userId"].(string))
		default:
			log.Println("Unknown message type:", messageType)
		}
		log.Println("Message processed")
	}
}

func main() {
	// runtime.GOMAXPROCS(4)
	lib.ConnectDB()
	r := mux.NewRouter()

	r.HandleFunc("/api/transaction", transaction.Handler)
	r.HandleFunc("/api/game-types", gametype.Handler)

	r.HandleFunc("/ws", handleWebSocket)
	api := r.PathPrefix("/api").Subrouter()
	userRouter := api.PathPrefix("/user").Subrouter()
	authRouter := api.PathPrefix("/auth").Subrouter()
	adminRouter := api.PathPrefix("/admin").Subrouter()

	user.Handler(userRouter)
	auth.Handler(authRouter)
	admin.Handler(adminRouter)

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{os.Getenv("FRONTEND_URL")}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Origin", "Content-Type", "Authorization"}),
	)(r)

	loggingHandler := handlers.LoggingHandler(os.Stdout, corsHandler)

	fmt.Println("WebSocket server listening on :8080")
	if err := http.ListenAndServe(":8080", loggingHandler); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}

// nodemon --exec go run main.go --signal SIGTERM
