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
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
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
		_, message, err := conn.ReadMessage()
		gameInstance := gameManager.GetInstance()
		if err != nil {
			targetUserId, exist := gameInstance.UserConnectionMap[conn]
			if exist {
				delete(gameInstance.UserConnectionMap, conn)
				gameInstance.DeleteUser(targetUserId)
				// gameInstance.GameQueue.Enqueue(gameInstance.Context, map[string]interface{}{
				// 	"type": "delete-user",
				// 	"data": map[string]string{
				// 		"userId": targetUserId,
				// 	},
				// })
			}
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
		log.Println("messageType", messageType)
		switch messageType {
		case "add-user":
			gameInstance.AddUser(messageData["userId"].(string), messageData["publicKey"].(string), conn)
		case "join-random-game":
			if lib.UnderMaintenance {
				targetUser, exist := gameInstance.GetUser(messageData["userId"].(string))
				if exist {
					targetUser.SendMessage("error", map[string]interface{}{
						"message": "We are under maintenance please try after some time",
					})
				}
			} else {
				log.Println("joining game", messageData["userId"].(string))
				gameInstance.GameQueue.Enqueue(gameInstance.Context, map[string]interface{}{
					"type": "join-game",
					"data": messageData,
				})
			}
		case "update-board":
			gameInstance.RedisClient.Publish(gameInstance.Context, messageData["gameId"].(string), string(message))
		case "game-over":
			messageData["pid"] = os.Getpid()
			payload, _ := json.Marshal(map[string]interface{}{
				"type": "game-over",
				"data": messageData,
			})
			gameInstance.RedisClient.Publish(gameInstance.Context, messageData["gameId"].(string), string(payload))
		default:
			log.Println("Unknown message type:", messageType)
		}
		log.Println("Message processed")
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	var wg sync.WaitGroup
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	gameManager.InitiateInstance(ctx, wg)

	defer gameManager.GetInstance().RedisClient.Close()

	lib.ConnectDB()
	r := mux.NewRouter()

	r.HandleFunc("/api/transaction", transaction.Handler)
	r.HandleFunc("/api/game-types", gametype.Handler)
	r.HandleFunc("/pid", func(w http.ResponseWriter, r *http.Request) {
		log.Println(os.Getppid())
	})

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

	server := &http.Server{
		Addr:    ":8080",
		Handler: loggingHandler,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		gameManager.GetInstance().SubscribeGame(ctx, "mari-arena-global")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("WebSocket server listening on :8080")
		if err := server.ListenAndServe(); err != nil {
			log.Fatal("ListenAndServe error:", err)
		}
	}()

	<-stop
	fmt.Println("Shutting down services...")

	cancel()

	serverCtx, serverCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer serverCancel()
	if err := server.Shutdown(serverCtx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	} else {
		fmt.Println("Server gracefully stopped")
	}

	wg.Wait()
	fmt.Println("All background tasks completed. Exiting.")
}

// nodemon --exec go run main.go --signal SIGTERM
