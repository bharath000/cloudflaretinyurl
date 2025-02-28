package redisqueue

import (
	"context"
	"fmt"
	"log"

	"cloudflare-tinyurl/rediscounter"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

const queueKey = "expired_click_queue" // Shared queue across instances

// Initialize Redis connection
func InitRedisQueue(redisClient *redis.Client) {
	rdb = redisClient
}

// Push an expired click event to the shared queue
func PushExpiredClick(snowflakeID string) {
	err := rdb.LPush(context.Background(), queueKey, snowflakeID).Err()
	if err != nil {
		log.Println("Failed to push to expired click queue:", err)
	}
}

// Processing expired keys
func ProcessExpiredClicks() {
	for {
		// BRPOP ensures only one instance processes each message
		result, err := rdb.BRPop(context.Background(), 0, queueKey).Result()
		if err != nil {
			log.Println("Error processing expired clicks:", err)
			continue
		}

		// Use the entire key (Snowflake key with "click:shortURL:snowflakeID")
		fullKey := result[1]
		log.Println("Processing expired click event for key:", fullKey)

		// Acquire lock to prevent race conditions
		lockKey := fmt.Sprintf("lock:%s", fullKey)
		lockAcquired := acquireLock(lockKey, 2)
		if !lockAcquired {
			log.Println("Skipping processing for:", fullKey, "Another instance is handling it.")
			continue
		}
		defer releaseLock(lockKey)

		// Decrement the counters using the full key
		err = rediscounter.DecrementGlobalCounter(fullKey)
		if err != nil {
			log.Println("Error decrementing counters for key:", fullKey, err)
		}
	}
}
