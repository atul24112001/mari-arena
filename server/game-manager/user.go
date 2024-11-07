package gameManager

import (
	"encoding/json"
	"log"

	"github.com/gofiber/contrib/websocket"
	// "github.com/gorilla/websocket"
)

type User struct {
	Id            string
	CurrentGameId string
	PublicKey     string
	Ws            *websocket.Conn
}

func (user *User) SendMessage(messageType string, data map[string]interface{}) {
	jsonByte, err := json.Marshal(map[string]interface{}{
		"type": messageType,
		"data": data,
	})

	if err != nil {
		log.Println("Error converting map into byte:", err)
	}

	if err := user.Ws.WriteMessage(int(1), jsonByte); err != nil {
		log.Println("Error writing message:", err)
	}
}
