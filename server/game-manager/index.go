package gameManager

import (
	"context"
	"errors"
	"flappy-bird-server/lib"
	"flappy-bird-server/model"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron"
	// "github.com/robfig/cron/v3"
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
	DbQueue      Queue
	GameQueue    Queue
}

var instance *GameManager
var once sync.Once

func GetInstance() *GameManager {
	once.Do(func() {
		redisAddr := os.Getenv("REDIS_URL")
		opt, _ := redis.ParseURL(redisAddr)
		dbClient := redis.NewClient(opt)
		gameClient := redis.NewClient(opt)
		dbQueue := Queue{
			client:        dbClient,
			queueName:     "mari-arena-db-queue",
			processingKey: "mari-arena-db-queue:processing",
			timeout:       10 * time.Second,
		}
		gameQueue := Queue{
			client:        gameClient,
			queueName:     "mari-arena-queue",
			processingKey: "mari-arena-queue:processing",
			timeout:       10 * time.Second,
		}
		instance = &GameManager{
			Users:        make(map[string]User),
			NewGame:      map[string]NewGame{},
			StartedGames: map[string]Game{},
			DbQueue:      dbQueue,
			GameQueue:    gameQueue,
		}
		ctx := context.Background()

		c := cron.New()
		c.AddFunc("@hourly", func() {
			dbQueue.RetryFailedTasks(ctx)
			gameQueue.RetryFailedTasks(ctx)
		})

		go dbQueue.ProcessQueue(ctx)
		go gameQueue.ProcessQueue(ctx)

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
	log.Println("Adding user")
	gameManger.Users[userId] = User{
		Id:        userId,
		Ws:        ws,
		PublicKey: publicKey,
	}
	log.Println("Added user")
}

func (gameManger *GameManager) CreateGame(maxUserCount int, winnerPrice int, entry int, gameTypeId string) (*Game, error) {
	newGameId, err := uuid.NewUUID()
	if err != nil {
		log.Println(err.Error())
		return &Game{}, errors.New("something went wrong while creating game id")
	}
	newGame := Game{
		Id:               newGameId.String(),
		Users:            make(map[string]bool),
		Status:           "staging",
		MaxUserCount:     maxUserCount,
		CurrentUserCount: 0,
		ScoreBoard:       make(map[string]Score),
		WinnerPrice:      winnerPrice,
		Entry:            entry,
	}

	item := map[string]interface{}{
		"type": "create-game",
		"data": map[string]interface{}{
			"id":           newGameId.String(),
			"entry":        entry,
			"winnerPrice":  winnerPrice,
			"gameTypeId":   gameTypeId,
			"maxUserCount": maxUserCount,
		},
	}

	err = gameManger.DbQueue.Enqueue(context.Background(), item)

	if err != nil {
		log.Println(err.Error())
		return &Game{}, errors.New("something went wrong while creating game")
	}

	return &newGame, nil
}

func (gameManger *GameManager) JoinGame(userId string, gameTypeId string) {
	targetUser, exist := gameManger.GetUser(userId)
	if !exist {
		return
	}
	newGameMap, newGameMapExist := gameManger.NewGame[gameTypeId]
	if !newGameMapExist {
		cacheGameTypeMap, cacheGameTypeMapExist := gameTypeMap[gameTypeId]

		if !cacheGameTypeMapExist || cacheGameTypeMap.LastUpdated+36000 < int(time.Now().Unix()) {
			var gameType model.GameType
			err := lib.Pool.QueryRow(context.Background(), `SELECT id, title, currency, "maxPlayer", winner, entry FROM public.gametypes WHERE id = $1`, gameTypeId).Scan(&gameType.Id, &gameType.Title, &gameType.Currency, &gameType.MaxPlayer, &gameType.Winner, &gameType.Entry)
			if err != nil {
				log.Println("105", err.Error())
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

		_newGame, err := gameManger.CreateGame(cacheGameTypeMap.GameType.MaxPlayer, cacheGameTypeMap.GameType.Winner, cacheGameTypeMap.GameType.Entry, cacheGameTypeMap.GameType.Id)
		if err != nil {
			log.Println("120", err.Error())
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
		log.Println("Game Created", gameTypeId, newGameMap)
	}

	newGame := newGameMap.Game
	log.Printf("Join game %s by user %s", newGame.Id, userId)
	if newGame.CurrentUserCount == newGame.MaxUserCount {
		log.Println("Game is full")
		return
	}

	var currentBalance int
	err := lib.Pool.QueryRow(context.Background(), `SELECT "solanaBalance"  FROM public.users WHERE id = $1`, userId).Scan(&currentBalance)

	if err != nil {
		log.Println("151", err.Error())
		targetUser.SendMessage("error", map[string]interface{}{
			"message": "Something went wrong while fetching current balance",
		})
		return
	}

	if currentBalance < newGame.Entry {
		log.Println("159", err.Error())
		targetUser.SendMessage("error", map[string]interface{}{
			"message": "Insufficient balance",
		})
		return
	}

	if !newGame.Users[userId] {
		err = gameManger.DbQueue.Enqueue(context.Background(), map[string]interface{}{
			"type": "add-participant",
			"data": map[string]interface{}{
				"userId": userId,
				"gameId": newGame.Id,
			},
		})

		if err != nil {
			targetUser.SendMessage("error", map[string]interface{}{
				"message": "Something went wrong while adding participant in db",
			})
			return
		}
	}

	keys := make([]string, 0, newGame.MaxUserCount)
	for k := range newGame.Users {
		participants, exist := gameManger.GetUser(k)
		if exist {
			keys = append(keys, k)
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

	newGameMap.Game = newGame
	gameManger.NewGame[gameTypeId] = newGameMap

	log.Printf("Game joined userId %s - gameId %s", userId, newGame.Id)
	log.Printf("current: %d - max: %d", newGame.CurrentUserCount, newGame.MaxUserCount)
	if newGame.CurrentUserCount == newGame.MaxUserCount {
		newGame.Status = "ongoing"
		gameManger.StartedGames[newGame.Id] = newGame
		delete(gameManger.NewGame, gameTypeId)
		keys = append(keys, userId)

		ids := ""
		for _, id := range keys {
			if ids == "" {
				ids += fmt.Sprintf(`'%s'`, id)
			} else {
				ids += fmt.Sprintf(`, '%s'`, id)
			}
		}

		err = gameManger.DbQueue.Enqueue(context.Background(), map[string]interface{}{
			"type": "start-game",
			"data": map[string]interface{}{
				"gameId": newGame.Id,
			},
		})
		if err != nil {
			for _, id := range keys {
				participant, exist := gameManger.GetUser(id)
				if exist {
					log.Println(err.Error())
					participant.SendMessage("error", map[string]interface{}{
						"message": "Something went wrong while updating game status",
					})
				}
			}
		}

		err := gameManger.DbQueue.Enqueue(context.Background(), map[string]interface{}{
			"type": "collect-entry",
			"data": map[string]interface{}{
				"ids":   ids,
				"entry": newGameMap.GameType.Entry,
			},
		})

		for _, id := range keys {
			participant, exist := gameManger.GetUser(id)
			if exist {
				if err != nil {
					log.Println(err.Error())
					participant.SendMessage("error", map[string]interface{}{
						"message": "Something went wrong while collecting entry fees",
					})
				} else {
					gameManger.Users[id] = User{
						Id:            participant.Id,
						CurrentGameId: newGame.Id,
						PublicKey:     participant.PublicKey,
						Ws:            participant.Ws,
					}
					participant.SendMessage("start-game", map[string]interface{}{})
				}
			}
		}
	}
}

func (gameManger *GameManager) DeleteUser(conn *websocket.Conn, targetUserId string) {
	if targetUserId != "" {
		targetUser, userExist := gameManger.GetUser(targetUserId)
		for gameId, g := range gameManger.StartedGames {
			_, existInGame := g.Users[targetUserId]
			if existInGame && userExist {
				log.Println("User exist in game", gameId, "currentGameId: ", targetUser.CurrentGameId, " end")
				if targetUser.CurrentGameId == "" {
					// user is in the game but game is not started yet
					inGameUsersMap := map[string]bool{}
					inGameUsersScoreboard := map[string]Score{}
					for key, _ := range g.Users {
						if key != targetUserId {
							inGameUsersMap[key] = true
							inGameUsersScoreboard[key] = g.ScoreBoard[key]
						}
					}
					gameManger.StartedGames[gameId] = Game{
						Id:               g.Id,
						MaxUserCount:     g.MaxUserCount,
						CurrentUserCount: g.CurrentUserCount - 1,
						WinnerPrice:      g.WinnerPrice,
						Entry:            g.Entry,
						Users:            inGameUsersMap,
						Status:           g.Status,
						ScoreBoard:       inGameUsersScoreboard,
					}
				} else {
					gameManger.GameOver(gameId, targetUserId)
				}
			}
		}
	}
	delete(gameManger.Users, targetUserId)
}

func (gameManger *GameManager) DeleteGame(gameId string) {
	delete(gameManger.StartedGames, gameId)
}

func (gameManager *GameManager) UpdateBoard(gameId string, userId string) {
	if _, exist := gameManager.GetUser(userId); !exist {
		return
	}
	targetGame := gameManager.GetGame(gameId)
	targetGame.UpdateScore(userId)

	oldBoard, exist := targetGame.ScoreBoard[userId]
	if exist {
		targetGame.ScoreBoard[userId] = Score{
			IsAlive: oldBoard.IsAlive,
			Points:  oldBoard.Points + 1,
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

	if alivePlayers == 0 {
		if winnerId != "" {
			err := gameManager.DbQueue.Enqueue(context.Background(), map[string]interface{}{
				"type": "end-game",
				"data": map[string]string{
					"gameId":   gameId,
					"winnerId": winnerId,
				},
			})

			if err != nil {
				newLine := fmt.Sprintf("ERROR_UPDATING_GAME-gameId_%s-status_%s-userId_%s-amount-%d\n", gameId, "completed", userId, targetGame.WinnerPrice)
				lib.ErrorLogger(newLine, "errors.txt")
				return
			}

			err = gameManager.DbQueue.Enqueue(context.Background(), map[string]interface{}{
				"type": "update-balance",
				"data": map[string]interface{}{
					"winnerId": winnerId,
					"amount":   targetGame.WinnerPrice,
				},
			})
			if err != nil {
				log.Println(err.Error())
				newLine := fmt.Sprintf("ERROR_UPDATING_USER_BALANCE-userId_%s-amount_%d\n", winnerId, targetGame.WinnerPrice)
				lib.ErrorLogger(newLine, "errors.txt")
				return
			}
		}

		for k, _ := range targetGame.Users {
			log.Println("370", k)
			participant, exist := gameManager.GetUser(k)
			if exist {
				gameManager.Users[k] = User{
					Id:            participant.Id,
					CurrentGameId: "",
					PublicKey:     participant.PublicKey,
					Ws:            participant.Ws,
				}
				if k == winnerId {
					participant.SendMessage("winner", map[string]interface{}{
						"amount": targetGame.WinnerPrice - targetGame.Entry,
					})
				} else {
					participant.SendMessage("loser", map[string]interface{}{
						"amount": targetGame.Entry,
					})
				}
			}
		}
		gameManager.DeleteGame(gameId)
	}
}
