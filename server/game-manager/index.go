package gameManager

import (
	"errors"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type GameManager struct {
	Users        map[string]User
	NewGame      Game
	StartedGames map[string]Game
}

var instance *GameManager
var once sync.Once

func GetInstance() *GameManager {
	once.Do(func() {
		instance = &GameManager{
			Users: make(map[string]User),
		}
	})
	return instance
}

func (gameManger *GameManager) GetGame(gameId string) *Game {
	targetGame := gameManger.StartedGames[gameId]
	return &targetGame
}

func (gameManger *GameManager) GetUser(userId string) *User {
	targetUser := gameManger.Users[userId]
	return &targetUser
}

func (gameManger *GameManager) AddUser(userId string, ws *websocket.Conn) {
	gameManger.Users[userId] = User{
		Id: userId,
		Ws: ws,
	}
}

func (gameManger *GameManager) CreateGame() (*Game, error) {
	gameId, err := uuid.NewRandom()
	if err != nil {
		return &Game{}, errors.New("Something went wrong while creating game")
	}

	return &Game{
		Id:               gameId.String(),
		Users:            make(map[string]bool),
		IsStarted:        false,
		MaxUserCount:     2,
		CurrentUserCount: 0,
		ScoreBoard:       make(map[string]Score),
	}, nil
}

func (gameManger *GameManager) JoinGame(userId string) {
	targetUser := gameManger.GetUser(userId)
	newGame := gameManger.NewGame

	if newGame.Id == "" {
		_newGame, err := gameManger.CreateGame()
		if err != nil {
			targetUser.SendMessage("error", map[string]interface{}{
				"message": "Error while creating new game",
			})
			return
		}
		newGame = *_newGame
	}

	keys := make([]string, 0, len(newGame.Users))
	for k := range newGame.Users {
		keys = append(keys, k)
	}

	for k := range newGame.Users {
		participants := gameManger.GetUser(k)
		participants.SendMessage("new-user", map[string]interface{}{
			"userId": userId,
			"gameId": newGame.Id,
		})
	}

	targetUser.SendMessage("join-game", map[string]interface{}{
		"users":  keys,
		"gameId": newGame.Id,
	})

	newGame.CurrentUserCount += 1
	gameManger.NewGame = newGame
	newGame.Users[userId] = true
	newGame.ScoreBoard[userId] = Score{
		IsAlive: true,
		Points:  0,
	}

	if newGame.CurrentUserCount == newGame.MaxUserCount {
		keys = append(keys, userId)
		newGame.IsStarted = true
		gameManger.StartedGames[newGame.Id] = newGame
		gameManger.NewGame = Game{}

		for _, id := range keys {
			participants := gameManger.GetUser(id)
			participants.SendMessage("start-game", map[string]interface{}{})
		}
	}
}

func (gameManger *GameManager) DeleteUser(conn *websocket.Conn) {
	newUsers := make(map[string]User)
	for key, user := range gameManger.Users {
		if user.Ws != conn {
			log.Println("Deleting", key)
			newUsers[key] = user
		}
	}
	gameManger.Users = newUsers
}

func (gameManger *GameManager) DeleteGame(gameId string) {
	delete(gameManger.StartedGames, gameId)
}

func (gameManager *GameManager) UpdateBoard(gameId string, userId string) {
	targetGame := gameManager.GetGame(gameId)
	targetGame.UpdateScore(userId)

	for k, _ := range targetGame.Users {
		if k != userId {
			targetUser := gameManager.GetUser(k)
			targetUser.SendMessage("update-board", map[string]interface{}{
				"scores": targetGame.ScoreBoard,
			})
		}
	}
}

func (gameManager *GameManager) GameOver(gameId string, userId string) {
	targetGame := gameManager.GetGame(gameId)
	targetGame.GameOver(userId)

	var alivePlayers = 0
	var winnerId = ""

	for k, _ := range targetGame.Users {
		if targetGame.ScoreBoard[k].IsAlive {
			alivePlayers += 1
			winnerId = k
		}
	}

	if alivePlayers == 1 {
		winner := gameManager.GetUser(winnerId)
		winner.SendMessage("winner", map[string]interface{}{})
	}
}
