package main

import (
	"encoding/json"
	gameManager "flappy-bird-server/game-manager"
	"flappy-bird-server/lib"
	"flappy-bird-server/transaction"
	"flappy-bird-server/user"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		log.Println(r.Host)
		if r.Host != "localhost:8080" && r.Host != "localhost:3000" {
			return false
		}
		return true
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade:", err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			gameManager.GetInstance().DeleteUser(conn)
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
			gameManager.GetInstance().AddUser(messageData["userId"].(string), conn)
		case "join-random-game":
			gameManager.GetInstance().JoinGame(messageData["userId"].(string))
		case "update-board":
			gameManager.GetInstance().UpdateBoard(messageData["gameId"].(string), messageData["userId"].(string))
		}
	}
}

func httpCorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		log.Println(origin)
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
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
	http.Handle("/api/transaction", httpCorsMiddleware(http.HandlerFunc(transaction.Handler)))

	fmt.Println("WebSocket server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}
