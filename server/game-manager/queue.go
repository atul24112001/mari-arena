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

// Enqueue adds a task to the queue
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

// Dequeue removes a task from the queue and moves it to the processing list
func (q *Queue) Dequeue(ctx context.Context) (string, error) {
	result, err := q.client.BRPopLPush(ctx, q.queueName, q.processingKey, 0).Result()
	if err == redis.Nil {
		return "", nil
	}
	return result, err
}

// Acknowledge removes a task from the processing list upon successful processing
func (q *Queue) Acknowledge(ctx context.Context, item string) error {
	return q.client.LRem(ctx, q.processingKey, 0, item).Err()
}

// RetryFailedTasks moves items back to the main queue if they exceed the processing timeout
func (q *Queue) RetryFailedTasks(ctx context.Context) error {
	items, err := q.client.LRange(ctx, q.processingKey, 0, -1).Result()
	if err != nil {
		return err
	}

	for _, item := range items {
		// Check the item's "processing time" with an arbitrary check using ZSET for better visibility if needed.
		// For now we simply dequeue to retry after processing time exceeded
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

// ProcessQueue continuously processes items and retries failed tasks periodically
func (q *Queue) ProcessQueue(ctx context.Context) {
	for {
		// Attempt to dequeue a task
		item, err := q.Dequeue(ctx)
		if err != nil {
			// Check for EOF (queue empty) and other errors
			if err == io.EOF {
				time.Sleep(2 * time.Second)
				continue
			} else {
				// Log and attempt reconnection if necessary
				log.Printf("Error dequeuing task: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}
		} else if item == "" {
			// If no items are in the queue, we wait and continue
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
			_, err = lib.Pool.Exec(context.Background(), `INSERT INTO public.games (id, "entryFee", "winningAmount", "gameTypeId", "maxPlayer")
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id, status,  "entryFee", "winningAmount", "maxPlayer"`, taskPayload["id"], taskPayload["entry"], taskPayload["winnerPrice"], taskPayload["gameTypeId"], taskPayload["maxUserCount"])
		case "add-participant":
			_, err = lib.Pool.Exec(context.Background(), `INSERT INTO public.participants ("userId", "gameId") VALUES ($1, $2)`, taskPayload["userId"], taskPayload["gameId"])
		case "start-game":
			_, err = lib.Pool.Exec(context.Background(), `UPDATE public.games SET status = $2 WHERE id =  $1`, taskPayload["gameId"], "ongoing")
		case "collect-entry":
			query := fmt.Sprintf(`UPDATE public.users SET "solanaBalance" = "solanaBalance" - $1 WHERE id IN (%s) AND "solanaBalance" >= $1`, taskPayload["ids"])
			_, err = lib.Pool.Exec(context.Background(), query, taskPayload["entry"])
		case "join-game":
			GetInstance().JoinGame(taskPayload["userId"].(string), taskPayload["gameTypeId"].(string))
		}

		if err != nil {
			// if err := q.RetryFailedTasks(ctx); err != nil {
			log.Printf("Failed to retry tasks: %s", err.Error())
			// }
		} else {
			if err := q.Acknowledge(ctx, item); err != nil {
				log.Fatalf("Failed to acknowledge task: %v", err)
			}

		}

	}
}

func Parse(jsonStr string, result interface{}) error {
	return json.Unmarshal([]byte(jsonStr), result)
}
