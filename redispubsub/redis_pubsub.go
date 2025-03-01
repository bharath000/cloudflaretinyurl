package redispubsub

import (
	"context"
	"log"
	"strings"

	"cloudflaretinyurl/redisqueue"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

// Initialize Redis Pub/Sub
func InitRedisPubSub(redisClient *redis.Client) {
	rdb = redisClient
}

// Listen for expired key events and push the full key (snowflake ID) to the queue
func ListenForExpiredClicks() {
	pubsub := rdb.PSubscribe(context.Background(), "__keyevent@0__:expired")

	for msg := range pubsub.Channel() {
		expiredKey := msg.Payload

		// Check if the expired key is a click event
		if strings.Contains(expiredKey, "click:") {
			log.Println("Detected expired click event for key:", expiredKey)

			// Push the full key (including snowflake ID) to the processing queue
			redisqueue.PushExpiredClick(expiredKey)
		}
	}
}
