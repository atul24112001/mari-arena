package main

import (
	"context"
	"encoding/json"
	"flappy-bird-server/admin"
	"flappy-bird-server/auth"
	gameManager "flappy-bird-server/game-manager"
	gametype "flappy-bird-server/game-type"
	"flappy-bird-server/lib"
	"flappy-bird-server/middleware"
	"flappy-bird-server/transaction"
	"flappy-bird-server/user"
	"fmt"
	"log"
	"net/http"
	"os"

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

	// var userId string
	for {
		n, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("err", err.Error())
			// gameManager.GetInstance().DeleteUser(conn, userId)
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

func httpCorsMiddleware(next http.Handler) http.Handler {
	return middleware.Logger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", os.Getenv("FRONTEND_URL"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	}))

}

func main() {
	// runtime.GOMAXPROCS(4)
	lib.ConnectDB()

	http.HandleFunc("/ws", handleWebSocket)
	http.Handle("/api/user", httpCorsMiddleware(http.HandlerFunc(user.Handler)))
	http.Handle("/api/auth", httpCorsMiddleware(http.HandlerFunc(auth.Handler)))
	http.Handle("/api/transaction", http.HandlerFunc(transaction.Handler))
	http.Handle("/api/game-types", httpCorsMiddleware(http.HandlerFunc(gametype.Handler)))
	http.Handle("/api/admin", httpCorsMiddleware(http.HandlerFunc(admin.Handler)))
	http.Handle("/api/admin/metric", httpCorsMiddleware(http.HandlerFunc(admin.GetMetrics)))
	http.Handle("/api/admin/maintenance", httpCorsMiddleware(http.HandlerFunc(admin.UpdateUnderMaintenance)))

	fmt.Println("WebSocket server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}

// nodemon --exec go run main.go --signal SIGTERM
