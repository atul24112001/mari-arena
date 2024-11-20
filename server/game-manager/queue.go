package gameManager

import (
	"context"
	"encoding/json"
	"flappy-bird-server/lib"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Queue struct {
	client        *redis.Client
	queueName     string
	processingKey string
	timeout       time.Duration
}

func (q *Queue) Enqueue(ctx context.Context, item map[string]interface{}) error {
	jsonData, err := json.Marshal(item)
	if err != nil {
		return err
	}
	err = q.client.LPush(ctx, q.queueName, string(jsonData)).Err()
	if err != nil {
		log.Println("Enqueue 29", err.Error())
		return err
	}
	return nil

}

func (q *Queue) Dequeue(ctx context.Context) (string, error) {
	result, err := q.client.BRPopLPush(ctx, q.queueName, q.processingKey, 0).Result()
	if err == redis.Nil {
		return "", nil
	}
	return result, err
}

func (q *Queue) Acknowledge(ctx context.Context, item string) error {
	return q.client.LRem(ctx, q.processingKey, 0, item).Err()
}

func (q *Queue) RetryFailedTasks(ctx context.Context) error {
	items, err := q.client.LRange(ctx, q.processingKey, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, item := range items {
		if time.Since(time.Now()) > q.timeout {
			if err := q.client.LPush(ctx, q.queueName, item).Err(); err != nil {
				return err
			}
			if err := q.client.LRem(ctx, q.processingKey, 0, item).Err(); err != nil {
				return err
			}
			fmt.Printf("Requeued item: %s\n", item)
		}
	}
	return nil
}

func (q *Queue) ProcessQueue(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping redis queue")
			return
		default:
			item, err := q.Dequeue(ctx)
			if err != nil {
				if err == io.EOF {
					time.Sleep(2 * time.Second)
					continue
				} else {
					log.Printf("Error dequeuing task: %v", err)
					time.Sleep(2 * time.Second)
					continue
				}
			} else if item == "" {
				time.Sleep(2 * time.Second)
				continue
			}

			var parsedData map[string]interface{}
			err = Parse(item, &parsedData)
			if err != nil {
				fmt.Printf("Error in Parse: %v\n", err)
				return
			}

			taskType := parsedData["type"].(string)
			taskPayload := parsedData["data"].(map[string]interface{})
			log.Printf("Processing ====== %s", taskType)
			switch taskType {
			case "create-game":
				err = CreateGame(ctx, taskPayload)
			case "add-participant":
				err = AddParticipant(ctx, taskPayload)
			case "start-game":
				err = StartGame(ctx, taskPayload)
			case "collect-entry":
				err = CollectEntry(ctx, taskPayload)
			case "join-game":
				JoinGame(ctx, taskPayload)
			case "end-game":
				err = EndGame(ctx, taskPayload)
			case "update-balance":
				err = UpdateBalance(ctx, taskPayload)
			case "delete-user":
				GetInstance().DeleteUser(taskPayload["userId"].(string))
			}

			if err != nil {
				log.Printf("Failed to retry tasks: %s", err.Error())
			} else {
				if err := q.Acknowledge(ctx, item); err != nil {
					log.Fatalf("Failed to acknowledge task: %v", err)
				}
			}
		}
	}
}

func Parse(jsonStr string, result interface{}) error {
	return json.Unmarshal([]byte(jsonStr), result)
}

func CreateGame(ctx context.Context, taskPayload map[string]interface{}) error {
	_, err := lib.Pool.Exec(ctx, `INSERT INTO public.games (id, "entryFee", "winningAmount", "gameTypeId", "maxPlayer")
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, status,  "entryFee", "winningAmount", "maxPlayer"`, taskPayload["id"], taskPayload["entry"], taskPayload["winnerPrice"], taskPayload["gameTypeId"], taskPayload["maxUserCount"])
	return err
}

func AddParticipant(ctx context.Context, taskPayload map[string]interface{}) error {
	_, err := lib.Pool.Exec(ctx, `INSERT INTO public.participants ("userId", "gameId") VALUES ($1, $2)`, taskPayload["userId"], taskPayload["gameId"])
	return err
}

func StartGame(ctx context.Context, taskPayload map[string]interface{}) error {
	_, err := lib.Pool.Exec(ctx, `UPDATE public.games SET status = $2 WHERE id =  $1`, taskPayload["gameId"], "ongoing")
	return err
}

func JoinGame(ctx context.Context, taskPayload map[string]interface{}) error {
	GetInstance().JoinGame(taskPayload["userId"].(string), taskPayload["gameTypeId"].(string))
	return nil
}

func EndGame(ctx context.Context, taskPayload map[string]interface{}) error {
	_, err := lib.Pool.Exec(ctx, `UPDATE public.games SET status = $2,  "winnerId" = $3 WHERE id = $1`, taskPayload["gameId"], "completed", taskPayload["winnerId"])
	return err
}

func CollectEntry(ctx context.Context, taskPayload map[string]interface{}) error {
	query := fmt.Sprintf(`UPDATE public.users SET "solanaBalance" = "solanaBalance" - $1 WHERE id IN (%s) AND "solanaBalance" >= $1`, taskPayload["ids"])
	_, err := lib.Pool.Exec(ctx, query, taskPayload["entry"])
	return err
}

func UpdateBalance(ctx context.Context, taskPayload map[string]interface{}) error {
	_, err := lib.Pool.Exec(ctx, `UPDATE public.users SET  "solanaBalance" = "solanaBalance"  + $2 WHERE id = $1`, taskPayload["winnerId"], taskPayload["amount"])
	return err
}
