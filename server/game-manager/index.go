package gameManager

import (
	"context"
	"errors"
	"flappy-bird-server/lib"
	"flappy-bird-server/model"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type NewGame struct {
	Game     Game
	GameType model.GameType
}

type GameTypeMap struct {
	LastUpdated int
	GameType    model.GameType
}

var gameTypeMap = map[string]GameTypeMap{}

type GameManager struct {
	Users        map[string]User
	NewGame      map[string]NewGame
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

func (gameManger *GameManager) GetUser(userId string) (*User, bool) {
	targetUser, exist := gameManger.Users[userId]
	return &targetUser, exist
}

func (gameManger *GameManager) AddUser(userId string, publicKey string, ws *websocket.Conn) {
	gameManger.Users[userId] = User{
		Id:        userId,
		Ws:        ws,
		PublicKey: publicKey,
	}
}

func (gameManger *GameManager) CreateGame(maxUserCount int, winnerPrice int, entry int) (*Game, error) {
	gameId, err := uuid.NewRandom()
	if err != nil {
		return &Game{}, errors.New("Something went wrong while creating game")
	}

	return &Game{
		Id:               gameId.String(),
		Users:            make(map[string]bool),
		IsStarted:        false,
		MaxUserCount:     maxUserCount,
		CurrentUserCount: 0,
		ScoreBoard:       make(map[string]Score),
		WinnerPrice:      winnerPrice,
		Entry:            entry,
	}, nil
}

func (gameManger *GameManager) JoinGame(userId string, gameTypeId string) {
	targetUser, _ := gameManger.GetUser(userId)
	newGameMap, newGameMapExist := gameManger.NewGame[gameTypeId]

	if !newGameMapExist {
		cacheGameTypeMap, cacheGameTypeMapExist := gameTypeMap[gameTypeId]

		if !cacheGameTypeMapExist || cacheGameTypeMap.LastUpdated+36000 < int(time.Now().Unix()) {
			var gameType model.GameType
			err := lib.Pool.QueryRow(context.Background(), `SELECT id, title, currency, "maxPlayer", winner, entry FROM public.gametypes WHERE id = $1`, gameTypeId).Scan(&gameType.Id, &gameType.Title, &gameType.Currency, &gameType.MaxPlayer, &gameType.Winner, &gameType.Entry)
			if err != nil {
				targetUser.SendMessage("error", map[string]interface{}{
					"message": "Invalid game type",
				})
				return
			}
			gameTypeMap[gameTypeId] = GameTypeMap{
				LastUpdated: int(time.Now().Unix()),
				GameType:    gameType,
			}
			cacheGameTypeMap = gameTypeMap[gameTypeId]
		}

		_newGame, err := gameManger.CreateGame(cacheGameTypeMap.GameType.MaxPlayer, cacheGameTypeMap.GameType.Winner, cacheGameTypeMap.GameType.Entry)
		if err != nil {
			targetUser.SendMessage("error", map[string]interface{}{
				"message": "Error while creating new game",
			})
			return
		}

		newGameMap = NewGame{
			Game:     *_newGame,
			GameType: cacheGameTypeMap.GameType,
		}
		gameManger.NewGame[gameTypeId] = newGameMap
	}

	newGame := newGameMap.Game
	keys := make([]string, 0, len(newGame.Users))
	for k := range newGame.Users {
		keys = append(keys, k)
	}

	for k := range newGame.Users {
		participants, exist := gameManger.GetUser(k)
		if exist {
			participants.SendMessage("new-user", map[string]interface{}{
				"userId": userId,
				"gameId": newGame.Id,
			})
		}
	}

	targetUser.SendMessage("join-game", map[string]interface{}{
		"users":  keys,
		"gameId": newGame.Id,
	})

	newGame.CurrentUserCount += 1
	newGame.Users[userId] = true
	newGame.ScoreBoard[userId] = Score{
		IsAlive: true,
		Points:  0,
	}

	gameManger.NewGame[gameTypeId] = newGameMap

	if newGame.CurrentUserCount == newGame.MaxUserCount {
		keys = append(keys, userId)
		newGame.IsStarted = true
		gameManger.StartedGames[newGame.Id] = newGame
		delete(gameManger.NewGame, gameTypeId)

		ids := ""
		for _, id := range keys {
			if ids == "" {
				ids += fmt.Sprintf(`'%s'`, id)
			} else {
				ids += fmt.Sprintf(`, '%s'`, id)
			}
		}

		lib.Pool.Exec(context.Background(), `INSERT INTO public.games (id, status, "winningAmount", type, gameTypeId, users)`, newGameMap.GameType.Entry)
		query := fmt.Sprintf(`UPDATE public.users SET "solanaBalance" = "solanaBalance" - $1 WHERE id IN (%s) AND "solanaBalance" >= $1`, ids)
		_, err := lib.Pool.Exec(context.Background(), query, newGameMap.GameType.Entry)

		for _, id := range keys {
			participant, exist := gameManger.GetUser(id)
			if exist {
				if err != nil {
					participant.SendMessage("error", map[string]interface{}{
						"message": "Something went wrong while collecting entry fees",
					})
				} else {
					participant.SendMessage("start-game", map[string]interface{}{})
				}
			}
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
			targetUser, exist := gameManager.GetUser(k)
			if exist {
				targetUser.SendMessage("update-board", map[string]interface{}{
					"scores": targetGame.ScoreBoard,
				})
			}
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

		}
	}

	var highestPoints = 0
	for k, s := range targetGame.ScoreBoard {
		if s.Points > highestPoints {
			highestPoints = s.Points
			winnerId = k
		}
	}

	if alivePlayers == 0 && winnerId != "" {
		winner, _ := gameManager.GetUser(winnerId)
		winner.SendMessage("winner", map[string]interface{}{})
		_, err := lib.Pool.Exec(context.Background(), "UPDATE public.games SET (status, winnerEmail) VALUES ($2, $3) WHERE id = $1", gameId, "completed", winner.Id)
		if err == nil {
			_, err := lib.Pool.Exec(context.Background(), `UPDATE public.users SET  "solanaBalance" = "solanaBalance"  + $2 WHERE id = $1`, userId, targetGame.WinnerPrice)
			if err == nil {
				gameManager.DeleteGame(gameId)
			}
		}
	}
}
