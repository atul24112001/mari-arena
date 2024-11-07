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

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
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
			Users:        make(map[string]User),
			NewGame:      map[string]NewGame{},
			StartedGames: map[string]Game{},
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

func (gameManger *GameManager) AddUser(userId string, publicKey string, ws *websocket.Conn) string {
	gameManger.Users[userId] = User{
		Id:        userId,
		Ws:        ws,
		PublicKey: publicKey,
	}
	return userId
}

func (gameManger *GameManager) CreateGame(maxUserCount int, winnerPrice int, entry int, gameTypeId string) (*Game, error) {
	newGameId, err := uuid.NewUUID()
	var newGame Game
	if err != nil {
		log.Println(err.Error())
		return &Game{}, errors.New("Something went wrong while creating game id")
	}

	err = lib.Pool.QueryRow(context.Background(), `INSERT INTO public.games (id, "entryFee", "winningAmount", "gameTypeId", "maxPlayer")
	 VALUES ($1, $2, $3, $4, $5)
	 RETURNING id, status,  "entryFee", "winningAmount", "maxPlayer"`, newGameId.String(), entry, winnerPrice, gameTypeId, maxUserCount).Scan(&newGame.Id, &newGame.Status, &newGame.Entry, &newGame.WinnerPrice, &newGame.MaxUserCount)

	if err != nil {
		log.Println(err.Error(), "78")
		return &Game{}, errors.New("Something went wrong while creating game")
	}

	return &Game{
		Id:               newGame.Id,
		Users:            make(map[string]bool),
		Status:           newGame.Status,
		MaxUserCount:     maxUserCount,
		CurrentUserCount: 0,
		ScoreBoard:       make(map[string]Score),
		WinnerPrice:      newGame.WinnerPrice,
		Entry:            newGame.Entry,
	}, nil
}

func (gameManger *GameManager) JoinGame(userId string, gameTypeId string) {
	targetUser, _ := gameManger.GetUser(userId)
	newGameMap, newGameMapExist := gameManger.NewGame[gameTypeId]
	log.Println(newGameMap.Game.Users)
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
	log.Println("Here 136")
	if newGame.CurrentUserCount == newGame.MaxUserCount {
		log.Println("Game is full")
		return
	}

	if newGame.Users[userId] {
		log.Println("User already exist")
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

	_, err = lib.Pool.Exec(context.Background(), `INSERT INTO public.participants ("userId", "gameId") VALUES ($1, $2)`, userId, newGame.Id)

	if err != nil {
		log.Println("178", err.Error())
		targetUser.SendMessage("error", map[string]interface{}{
			"message": "Something went wrong while adding participant in db",
		})
		return
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

	log.Println("Game joined " + userId + " " + newGame.Id)

	newGameMap.Game = newGame
	gameManger.NewGame[gameTypeId] = newGameMap

	log.Println(newGame.CurrentUserCount, newGame.MaxUserCount)
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

		_, err := lib.Pool.Exec(context.Background(), `UPDATE public.games SET status = $2 WHERE id =  $1`, newGame.Id, "ongoing")
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

		query := fmt.Sprintf(`UPDATE public.users SET "solanaBalance" = "solanaBalance" - $1 WHERE id IN (%s) AND "solanaBalance" >= $1`, ids)
		_, err = lib.Pool.Exec(context.Background(), query, newGameMap.GameType.Entry)

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
	// targetString := fmt.Sprintf("%s update score by 1 in %s", userId, gameId)
	// hash := lib.HashString(targetString)

	// if hash != token {
	// 	targetUser, userExist := gameManager.GetUser(userId)
	// 	if userExist {
	// 		targetUser.SendMessage("error", map[string]interface{}{
	// 			"message": "Invalid request",
	// 		})
	// 	}
	// 	return
	// }
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
	log.Println("Game over", userId)
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
			log.Println("346 gameId", targetGame)
			_, err := lib.Pool.Exec(context.Background(), `UPDATE public.games SET status = $2,  "winnerId" = $3 WHERE id = $1`, gameId, "completed", winnerId)

			if err != nil {
				log.Println(err.Error())
				newLine := fmt.Sprintf("ERROR_UPDATING_GAME-gameId_%s-status_%s-userId_%s-amount-%d\n", gameId, "completed", userId, targetGame.WinnerPrice)
				lib.ErrorLogger(newLine, "errors.txt")
				return
			}

			log.Println("353")
			_, err = lib.Pool.Exec(context.Background(), `UPDATE public.users SET  "solanaBalance" = "solanaBalance"  + $2 WHERE id = $1`, winnerId, targetGame.WinnerPrice)
			if err != nil {
				log.Println(err.Error())
				newLine := fmt.Sprintf("ERROR_UPDATING_USER_BALANCE-userId_%s-amount_%d\n", winnerId, targetGame.WinnerPrice)
				lib.ErrorLogger(newLine, "errors.txt")
				return
			}
		}

		log.Println("367", alivePlayers, winnerId)
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

func (gameManager *GameManager) GameOver2(gameId string, userId string) {
	targetGame := gameManager.GetGame(gameId)
	targetGame.GameOver(userId)

	var alivePlayers = 0
	var winnerId = ""

	// Count alive players and track the player with the highest score among those alive
	var highestPoints = 0
	for k, user := range targetGame.ScoreBoard {
		if user.IsAlive {
			alivePlayers += 1
			// Only update winnerId if the player is alive and has the highest points
			if user.Points > highestPoints {
				highestPoints = user.Points
				winnerId = k
			}
		}
	}

	log.Println("Alive players:", alivePlayers, "Winner ID:", winnerId)

	// Proceed only if no players are alive, meaning the game is truly over
	if alivePlayers == 0 {
		if winnerId != "" {
			newBalance := 0

			// Update game status and winner
			_, err := lib.Pool.Exec(
				context.Background(),
				"UPDATE public.games SET status = $2, winnerEmail = $3 WHERE id = $1",
				gameId, "completed", winnerId,
			)
			if err != nil {
				newLine := fmt.Sprintf("ERROR_UPDATING_GAME-gameId_%s-status_%s\n", gameId, "completed")
				lib.ErrorLogger(newLine, "errors.txt")
				return
			}

			// Update winner's balance
			err = lib.Pool.QueryRow(
				context.Background(),
				`UPDATE public.users SET "solanaBalance" = "solanaBalance" + $2 WHERE id = $1 RETURNING "solanaBalance"`,
				winnerId, targetGame.WinnerPrice,
			).Scan(&newBalance)
			if err != nil {
				newLine := fmt.Sprintf("ERROR_UPDATING_USER_BALANCE-userId_%s-amount_%d\n", winnerId, targetGame.WinnerPrice)
				lib.ErrorLogger(newLine, "errors.txt")
			} else {
				gameManager.DeleteGame(gameId)
			}
		}

		// Notify all players about the game result
		for k := range targetGame.Users {
			participant, exist := gameManager.GetUser(k)
			if exist {
				gameManager.Users[k] = User{
					Id:            participant.Id,
					CurrentGameId: "",
					PublicKey:     participant.PublicKey,
					Ws:            participant.Ws,
				}
				if k == winnerId && winnerId != "" {
					participant.SendMessage("winner", map[string]interface{}{
						"amount": targetGame.WinnerPrice,
					})
				} else {
					participant.SendMessage("loser", map[string]interface{}{
						"amount": targetGame.Entry,
					})
				}
			}
		}
	}
	log.Println("Game over complete. Alive players:", alivePlayers, "Winner ID:", winnerId)
}
