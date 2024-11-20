package gameManager

import (
	"context"
	"encoding/json"
	"errors"
	"flappy-bird-server/lib"
	"flappy-bird-server/model"
	"fmt"
	"log"
	"os"
	"strconv"
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
	UserConnectionMap map[*websocket.Conn]string
	Users             map[string]User
	Subscriptions     map[string]bool
	StartedGames      map[string]Game
	DbQueue           Queue
	GameQueue         Queue
	RedisClient       *redis.Client
	Context           context.Context
}

type RedisGame struct {
	Game     Game
	Response map[string]string
}

var instance *GameManager
var once sync.Once

func InitiateInstance(ctx context.Context, wg sync.WaitGroup) {
	once.Do(func() {
		// client := redis.NewClient(&redis.Options{
		// 	Addr:     os.Getenv("REDIS_ADDRESS"),
		// 	Password: os.Getenv("REDIS_PASSWORD"),
		// 	DB:       0,
		// })

		opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
		if err != nil {
			log.Fatal("Error parsing redis url: ", err.Error())
			return
		}

		client := redis.NewClient(opt)
		dbQueue := Queue{
			client:        client,
			queueName:     "mari-arena-db-queue",
			processingKey: "mari-arena-db-queue:processing",
			timeout:       10 * time.Second,
		}
		gameQueue := Queue{
			client:        client,
			queueName:     "mari-arena-queue",
			processingKey: "mari-arena-queue:processing",
			timeout:       10 * time.Second,
		}
		instance = &GameManager{
			UserConnectionMap: make(map[*websocket.Conn]string),
			Users:             make(map[string]User),
			StartedGames:      map[string]Game{},
			DbQueue:           dbQueue,
			GameQueue:         gameQueue,
			RedisClient:       client,
			Subscriptions:     map[string]bool{},
			Context:           ctx,
		}

		for i := 0; i < 3; i++ {
			log.Println("Checking redis connection")
			r := client.Ping(ctx)
			if r.Err() == nil {
				log.Println("Redis connected successfully")
				break
			}
			if r.Err() != nil && i > 1 {
				log.Fatal("Error connecting redis: ", r.Err().Error())
				time.Sleep(1 * time.Second)
			}
		}

		c := cron.New()
		c.AddFunc("@hourly", func() {
			log.Println("Retrying failed tasks")
			dbQueue.RetryFailedTasks(ctx)
			gameQueue.RetryFailedTasks(ctx)
		})

		wg.Add(1)
		go func() {
			defer wg.Done()
			dbQueue.ProcessQueue(ctx)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			gameQueue.ProcessQueue(ctx)
		}()
	})
}

func GetInstance() *GameManager {
	return instance
}

func (gameManager *GameManager) GetGame(gameId string) (*Game, bool) {
	targetGame, exist := gameManager.StartedGames[gameId]
	return &targetGame, exist
}

func (gameManager *GameManager) GetUser(userId string) (*User, bool) {
	targetUser, exist := gameManager.Users[userId]
	return &targetUser, exist
}

func (gameManager *GameManager) AddUser(userId string, publicKey string, ws *websocket.Conn) {
	gameManager.UserConnectionMap[ws] = userId
	gameManager.Users[userId] = User{
		Id:        userId,
		Ws:        ws,
		PublicKey: publicKey,
	}
}

func (gameManager *GameManager) CreateGame(maxUserCount int, winnerPrice int, entry int, gameTypeId string) (*Game, error) {
	newGameId, err := uuid.NewUUID()
	if err != nil {
		log.Println(err.Error())
		return &Game{}, errors.New("something went wrong while creating game id")
	}
	newGame := Game{
		Id:               newGameId.String(),
		GameTypeId:       gameTypeId,
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

	err = gameManager.DbQueue.Enqueue(gameManager.Context, item)

	if err != nil {
		log.Println(err.Error())
		return &Game{}, errors.New("something went wrong while creating game")
	}

	return &newGame, nil
}

func (gameManager *GameManager) GetBalance(userId string) (int, error) {
	red := gameManager.RedisClient.Get(gameManager.Context, fmt.Sprintf("mr-balance-%s", userId))
	err := red.Err()
	balance := 0
	if err == nil {
		balance, err = strconv.Atoi(red.Val())
	} else {
		err = lib.Pool.QueryRow(gameManager.Context, `SELECT "solanaBalance"  FROM public.users WHERE id = $1`, userId).Scan(&balance)
	}
	return balance, err
}

func (gameManager *GameManager) SetBalance(userId string, amount int) error {
	red := gameManager.RedisClient.Set(gameManager.Context, fmt.Sprintf("mr-balance-%s", userId), amount, 24*time.Hour)
	return red.Err()
}

func (gameManager *GameManager) GetStagingGameFromRedis(gameKey string) (Game, error) {
	cmdReturn := gameManager.RedisClient.Get(gameManager.Context, gameKey)
	response, err := cmdReturn.Result()

	if err == nil {
		var game Game
		err = Parse(response, &game)

		if err == nil {
			return game, nil
		}
	}
	return Game{}, err
}

func (gameManager *GameManager) JoinGame(userId string, gameTypeId string) {
	// targetUser, exist := gameManager.GetUser(userId)
	// if !exist {
	// 	return
	// }
	gameKey := fmt.Sprintf("newGame:%s", gameTypeId)
	newGame, err := gameManager.GetStagingGameFromRedis(gameKey)

	if err != nil {
		cacheGameTypeMap, cacheGameTypeMapExist := gameTypeMap[gameTypeId]

		if !cacheGameTypeMapExist || cacheGameTypeMap.LastUpdated+36000 < int(time.Now().Unix()) {
			var gameType model.GameType
			err := lib.Pool.QueryRow(gameManager.Context, `SELECT id, title, currency, "maxPlayer", winner, entry FROM public.gametypes WHERE id = $1`, gameTypeId).Scan(&gameType.Id, &gameType.Title, &gameType.Currency, &gameType.MaxPlayer, &gameType.Winner, &gameType.Entry)
			if err != nil {
				gameManager.RedisClient.Publish(gameManager.Context, "mari-arena-global", lib.Stringify(map[string]interface{}{
					"type": "user-error",
					"data": map[string]string{
						"userId":  userId,
						"message": "Invalid game type",
					},
				}))
				// targetUser.SendMessage("error", map[string]interface{}{
				// 	"message": "Invalid game type",
				// })
				return
			}
			gameTypeMap[gameTypeId] = GameTypeMap{
				LastUpdated: int(time.Now().Unix()),
				GameType:    gameType,
			}
			cacheGameTypeMap = gameTypeMap[gameTypeId]
		}

		_newGame, err := gameManager.CreateGame(cacheGameTypeMap.GameType.MaxPlayer, cacheGameTypeMap.GameType.Winner, cacheGameTypeMap.GameType.Entry, cacheGameTypeMap.GameType.Id)
		if err != nil {
			gameManager.RedisClient.Publish(gameManager.Context, "mari-arena-global", lib.Stringify(map[string]interface{}{
				"type": "user-error",
				"data": map[string]string{
					"userId":  userId,
					"message": "Error while creating new game",
				},
			}))
			// targetUser.SendMessage("error", map[string]interface{}{
			// 	"message": "Error while creating new game",
			// })
			return
		}
		newGame = *_newGame
		payload, err := json.Marshal(newGame)
		if err == nil {
			cmd := gameManager.RedisClient.Set(gameManager.Context, gameKey, string(payload), 365*24*time.Hour)
			err = cmd.Err()
		}

		if err != nil {
			gameManager.RedisClient.Publish(gameManager.Context, "mari-arena-global", lib.Stringify(map[string]interface{}{
				"type": "user-error",
				"data": map[string]string{
					"userId":  userId,
					"message": "Error setting message",
				},
			}))
			// targetUser.SendMessage("error", map[string]interface{}{
			// 	"message": "Error setting message",
			// })
			return
		}
	}

	log.Printf("Join game %s by user %s", newGame.Id, userId)
	if newGame.CurrentUserCount == newGame.MaxUserCount {
		log.Println("Game is full")
		return
	}

	currentBalance, err := gameManager.GetBalance(userId)

	if err != nil {
		// log.Println("151", err.Error())
		gameManager.RedisClient.Publish(gameManager.Context, "mari-arena-global", lib.Stringify(map[string]interface{}{
			"type": "user-error",
			"data": map[string]string{
				"userId":  userId,
				"message": "Something went wrong while fetching current balance",
			},
		}))
		// targetUser.SendMessage("error", map[string]interface{}{
		// 	"message": "Something went wrong while fetching current balance",
		// })
		return

	}

	if currentBalance < newGame.Entry {
		gameManager.RedisClient.Publish(gameManager.Context, "mari-arena-global", lib.Stringify(map[string]interface{}{
			"type": "user-error",
			"data": map[string]string{
				"userId":  userId,
				"message": "Insufficient balance",
			},
		}))
		// targetUser.SendMessage("error", map[string]interface{}{
		// 	"message": "Insufficient balance",
		// })
		return
	}

	if !newGame.Users[userId] {
		err := gameManager.DbQueue.Enqueue(gameManager.Context, map[string]interface{}{
			"type": "add-participant",
			"data": map[string]interface{}{
				"userId": userId,
				"gameId": newGame.Id,
			},
		})

		if err != nil {
			gameManager.RedisClient.Publish(gameManager.Context, "mari-arena-global", lib.Stringify(map[string]interface{}{
				"type": "user-error",
				"data": map[string]string{
					"userId":  userId,
					"message": "Something went wrong while adding participant in db",
				},
			}))
			// targetUser.SendMessage("error", map[string]interface{}{
			// 	"message": "Something went wrong while adding participant in db",
			// })
			return
		}
	}

	keys := make([]string, 0, newGame.MaxUserCount)
	for k := range newGame.Users {
		participants, exist := gameManager.GetUser(k)
		if exist {
			keys = append(keys, k)
			participants.SendMessage("new-user", map[string]interface{}{
				"userId": userId,
				"gameId": newGame.Id,
			})
		}
	}

	newGame.CurrentUserCount += 1

	newGame.Users[userId] = true
	newGame.ScoreBoard[userId] = Score{
		IsAlive: true,
		Points:  0,
	}

	gameManager.RedisClient.Publish(gameManager.Context, "mari-arena-global", lib.Stringify(map[string]interface{}{
		"type": "user-join-game",
		"data": map[string]interface{}{
			"userId": userId,
			"users":  keys,
			"gameId": newGame.Id,
		},
	}))
	// targetUser.SendMessage("join-game", map[string]interface{}{
	// 	"users":  keys,
	// 	"gameId": newGame.Id,
	// })
	// gameManager.Users[targetUser.Id] = User{
	// 	Id:            targetUser.Id,
	// 	CurrentGameId: newGame.Id,
	// 	PublicKey:     targetUser.PublicKey,
	// 	Ws:            targetUser.Ws,
	// }

	// if !gameManager.Subscriptions[newGame.Id] {
	// 	gameManager.Subscriptions[newGame.Id] = true
	// 	go gameManager.SubscribeGame(gameManager.Context, newGame.Id)
	// }

	if newGame.CurrentUserCount == newGame.MaxUserCount {
		cmd := gameManager.RedisClient.Ping(gameManager.Context)
		if cmd.Err() == nil {
			keys = append(keys, userId)
			ids := ""
			for _, id := range keys {
				balance, err := gameManager.GetBalance(id)
				if err != nil {
					gameManager.SetBalance(id, balance-newGame.Entry)
				}
				if ids == "" {
					ids += fmt.Sprintf(`'%s'`, id)
				} else {
					ids += fmt.Sprintf(`, '%s'`, id)
				}
			}

			err = gameManager.DbQueue.Enqueue(gameManager.Context, map[string]interface{}{
				"type": "start-game",
				"data": map[string]interface{}{
					"gameId": newGame.Id,
				},
			})
			if err == nil {
				err = gameManager.DbQueue.Enqueue(gameManager.Context, map[string]interface{}{
					"type": "collect-entry",
					"data": map[string]interface{}{
						"ids":   ids,
						"entry": newGame.Entry,
					},
				})

				if err == nil {
					redisCmd := gameManager.RedisClient.Del(gameManager.Context, gameKey)
					err = redisCmd.Err()
					if err == nil {
						jsonString, err := json.Marshal(map[string]interface{}{
							"type": "start-game",
							"data": newGame,
						})
						if err == nil {
							redisCmd = gameManager.RedisClient.Publish(gameManager.Context, newGame.Id, string(jsonString))
							err = redisCmd.Err()
							log.Println(err)
						}
					}
				}
			}
		}
		if err != nil {
			errorPayload, err := json.Marshal(map[string]interface{}{
				"type": "error-starting-game",
				"data": newGame,
			})
			if err != nil {
				gameManager.RedisClient.Publish(gameManager.Context, newGame.Id, string(errorPayload))
			}

		}
	} else {
		payload, err := json.Marshal(newGame)
		if err != nil {
			gameManager.RedisClient.Publish(gameManager.Context, "mari-arena-global", lib.Stringify(map[string]interface{}{
				"type": "user-error",
				"data": map[string]string{
					"userId":  userId,
					"message": "Error while setting new game",
				},
			}))
			// targetUser.SendMessage("error", map[string]interface{}{
			// 	"message": "Error while setting new game",
			// })
			return
		}
		cmd := gameManager.RedisClient.Set(gameManager.Context, gameKey, string(payload), 365*24*time.Hour)

		if cmd.Err() != nil {
			gameManager.RedisClient.Publish(gameManager.Context, "mari-arena-global", lib.Stringify(map[string]interface{}{
				"type": "user-error",
				"data": map[string]string{
					"userId":  userId,
					"message": "Error while setting new game",
				},
			}))
			// targetUser.SendMessage("error", map[string]interface{}{
			// 	"message": "Error while setting new game",
			// })
			return
		}
	}
}

func (gameManager *GameManager) DeleteUser(targetUserId string) {
	targetUser, userExist := gameManager.GetUser(targetUserId)
	if userExist {
		log.Println("Deleting user: ", targetUserId)
		delete(gameManager.Users, targetUserId)
		if targetUser.CurrentGameId != "" {
			targetGame, gameExist := gameManager.GetGame(targetUser.CurrentGameId)
			if gameExist {
				if targetGame.Status == "ongoing" && targetGame.ScoreBoard[targetUserId].IsAlive {
					gameManager.GameOver(targetGame.Id, targetUserId)
				} else if targetGame.Status == "staging" {
					gameKey := fmt.Sprintf("newGame:%s", targetGame.GameTypeId)
					inGameUsers := targetGame.Users
					inGameScoreboard := targetGame.ScoreBoard

					delete(inGameUsers, targetUserId)
					delete(inGameScoreboard, targetUserId)

					updatedGame := Game{
						Id:               targetGame.Id,
						GameTypeId:       targetGame.GameTypeId,
						MaxUserCount:     targetGame.MaxUserCount,
						CurrentUserCount: targetGame.CurrentUserCount - 1,
						WinnerPrice:      targetGame.WinnerPrice,
						Entry:            targetGame.Entry,
						Users:            inGameUsers,
						Status:           "staging",
						ScoreBoard:       inGameScoreboard,
					}

					gameManager.StartedGames[targetGame.Id] = updatedGame
					payload, err := json.Marshal(updatedGame)
					if err == nil {
						gameManager.RedisClient.Set(gameManager.Context, gameKey, string(payload), 365*24*time.Hour)
					} else {
						gameManager.RedisClient.Del(gameManager.Context, gameKey)
					}
				}
			}
		}
	}
}

func (gameManager *GameManager) DeleteGame(gameId string) {
	delete(gameManager.StartedGames, gameId)
}

func (gameManager *GameManager) UpdateBoard(gameId string, userId string) {
	// if _, exist := gameManager.GetUser(userId); !exist {
	// 	return
	// }
	targetGame, gameExist := gameManager.GetGame(gameId)
	if gameExist {
		targetGame.UpdateScore(userId)
		// oldBoard, exist := targetGame.ScoreBoard[userId]
		// if exist {
		// 	targetGame.ScoreBoard[userId] = Score{
		// 		IsAlive: oldBoard.IsAlive,
		// 		Points:  oldBoard.Points + 1,
		// 	}
		// }
	}
}

func (gameManager *GameManager) GameOver(gameId string, userId string) {
	targetGame, gameExist := gameManager.GetGame(gameId)
	if gameExist {
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
			for k := range targetGame.Users {
				participant, exist := gameManager.GetUser(k)
				if exist {
					gameManager.Users[k] = User{
						Id:            participant.Id,
						CurrentGameId: "",
						PublicKey:     participant.PublicKey,
						Ws:            participant.Ws,
					}
					if k == winnerId {
						err := gameManager.DbQueue.Enqueue(gameManager.Context, map[string]interface{}{
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
						if winnerId != "" {
							err = gameManager.DbQueue.Enqueue(gameManager.Context, map[string]interface{}{
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
						balance, err := gameManager.GetBalance(winnerId)
						if err != nil {
							gameManager.SetBalance(winnerId, balance+targetGame.WinnerPrice)
						}
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
}
