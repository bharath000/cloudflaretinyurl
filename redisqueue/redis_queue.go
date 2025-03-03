package redisqueue

import (
	"context"
	"fmt"
	"log"

	"cloudflaretinyurl/rediscounter"
	"cloudflaretinyurl/redislocks"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

const queueKey = "expired_click_queue" // Shared queue across instances

// Initialize Redis Queue
func InitRedisQueue(redisClient *redis.Client) {
	rdb = redisClient
}

// Push an expired click event to the shared queue
func PushExpiredClick(fullKey string) {
	err := rdb.LPush(context.Background(), queueKey, fullKey).Err()
	if err != nil {
		log.Println("Failed to push to expired click queue:", err)
	}
}

// Process expired click events and decrement counters
func ProcessExpiredClicks() {
	for {
		// BRPOP ensures only one instance processes each message
		result, err := rdb.BRPop(context.Background(), 0, queueKey).Result()
		if err != nil {
			log.Println("Error processing expired clicks:", err)
			continue
		}

		fullKey := result[1]
		log.Println("Processing expired click event for key:", fullKey)

		// Acquire lock to prevent race conditions
		lockKey := fmt.Sprintf("lock:%s", fullKey)
		if redislocks.AcquireLock(lockKey, 2) {
			// Decrement the counters safely inside the lock scope
			err = rediscounter.DecrementGlobalCounter(fullKey)
			if err != nil {
				log.Println("Error decrementing counters for key:", fullKey, err)
			}

			// Release lock after processing
			redislocks.ReleaseLock(lockKey)
		} else {
			log.Println("Skipping processing for:", fullKey, "Another instance is handling it.")
		}
	}
}
