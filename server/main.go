package main

import (
	"encoding/json"
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
	"runtime"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("origin")
		log.Println("Origin", origin)
		return origin == os.Getenv("FRONTEND_URL")
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade:", err)
		return
	}

	var userId string
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
			userId = gameManager.GetInstance().AddUser(messageData["userId"].(string), messageData["publicKey"].(string), conn)
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

	http.HandleFunc("/ws", handleWebSocket)
	http.Handle("/api/user", httpCorsMiddleware(http.HandlerFunc(user.Handler)))
	http.Handle("/api/auth", httpCorsMiddleware(http.HandlerFunc(auth.Handler)))
	http.Handle("/api/transaction", http.HandlerFunc(transaction.Handler))
	http.Handle("/api/game-types", httpCorsMiddleware(http.HandlerFunc(gametype.Handler)))

	fmt.Println("WebSocket server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}
