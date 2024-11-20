package gameManager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

func (gameManager *GameManager) SubscribeGame(ctx context.Context, channel string) {
	pubsub := gameManager.RedisClient.Subscribe(ctx, channel)
	defer pubsub.Close()

	_, err := pubsub.Receive(ctx)
	if err != nil {
		log.Fatalf("Could not subscribe to channel: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stopping Redis Pub/Sub...")
			return
		case msg, ok := <-pubsub.Channel():
			var parsedData map[string]interface{}
			if !ok {
				fmt.Println("Redis Pub/Sub channel closed")
				return
			}
			err = Parse(msg.Payload, &parsedData)
			if err != nil {
				fmt.Printf("Error in Parse: %v\n", err)
				return
			}

			taskType := parsedData["type"].(string)
			taskPayload := parsedData["data"].(map[string]interface{})
			payloadString, err := json.Marshal(taskPayload)
			if err != nil {
				fmt.Printf("Error in Parse: %v\n", err)
				return
			}

			log.Println("44", taskType, taskPayload["users"])
			switch taskType {
			case "user-join-game":
				gameManager.UserJoinGame(taskPayload["userId"].(string), taskPayload["gameId"].(string), taskPayload["users"])
			case "user-error":
				gameManager.UserSendError(taskPayload["userId"].(string), taskPayload["message"].(string))
			case "start-game":
				gameManager.StartGame(string(payloadString))
			case "error-starting-game":
				gameManager.ErrorStatingGame(string(payloadString))
			case "update-board":
				gameManager.UpdateBoard(channel, taskPayload["userId"].(string))
			case "game-over":
				gameManager.GameOver(channel, taskPayload["userId"].(string))
			}
		}
	}
}

func (gameManager *GameManager) UserSendError(userId string, message string) {
	targetUser, useExist := gameManager.GetUser(userId)
	if !useExist {
		return
	}
	targetUser.SendMessage("error", map[string]interface{}{
		"message": message,
	})
}

func (gameManager *GameManager) UserJoinGame(userId string, gameId string, keys interface{}) {
	targetUser, useExist := gameManager.GetUser(userId)
	if !useExist {
		return
	}
	targetUser.SendMessage("join-game", map[string]interface{}{
		"users":  keys,
		"gameId": gameId,
	})
	gameManager.Users[targetUser.Id] = User{
		Id:            targetUser.Id,
		CurrentGameId: gameId,
		PublicKey:     targetUser.PublicKey,
		Ws:            targetUser.Ws,
	}

	if !gameManager.Subscriptions[gameId] {
		gameManager.Subscriptions[gameId] = true
		go gameManager.SubscribeGame(gameManager.Context, gameId)
	}

}

func (gameManager *GameManager) ErrorStatingGame(gameJsonString string) {
	var game Game
	err := Parse(gameJsonString, game)
	if err != nil {
		log.Println(gameJsonString)
		log.Fatal("Something went wrong while parsing game string")
		return
	}
	for k, _ := range game.Users {
		user, exist := gameManager.GetUser(k)
		if exist {
			user.SendMessage("error", map[string]interface{}{
				"message": "Error starting game",
			})
		}
	}

}

func (gameManager *GameManager) StartGame(gameJsonString string) {
	var game Game
	err := Parse(gameJsonString, &game)
	if err != nil {
		log.Println(gameJsonString)
		log.Fatal("Something went wrong while parsing game string", err.Error())
		return
	}
	users := []string{}
	for k, _ := range game.Users {
		_, exist := gameManager.GetUser(k)
		if exist {
			users = append(users, k)
		}
	}

	if len(users) == 0 {
		return
	}

	if err != nil {
		for _, k := range users {
			user, exist := gameManager.GetUser(k)
			if exist {
				user.SendMessage("error", map[string]interface{}{
					"message": "Something went wrong while starting game",
				})
			}
		}
		return
	}

	game.Status = "ongoing"
	gameManager.StartedGames[game.Id] = game

	for _, id := range users {
		participant, exist := gameManager.GetUser(id)
		if exist {
			if err != nil {
				log.Println(err.Error())
				participant.SendMessage("error", map[string]interface{}{
					"message": "Something went wrong while collecting entry fees",
				})
			} else {
				participant.SendMessage("start-game", map[string]interface{}{})
			}
		}
	}
}

// func
// gameManager.GetInstance().UpdateBoard(messageData["gameId"].(string), messageData["userId"].(string))
// gameManager.GetInstance().GameOver(messageData["gameId"].(string), messageData["userId"].(string))
