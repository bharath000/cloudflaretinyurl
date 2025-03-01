package rediscounter

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloudflaretinyurl/utils"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func InitRedisCounter(redisClient *redis.Client) {
	rdb = redisClient
}

// Extracts shortURL from Snowflake ID and updates global counters
func UpdateGlobalCounter(snowflakeID string) {
	ctx := context.Background()

	// Extract shortURL from the Snowflake ID
	shortURL, err := utils.DecodeShortURLFromSnowflakeID(snowflakeID)
	if err != nil {
		log.Println("Error extracting shortURL from Snowflake ID:", err)
		return
	}

	// Generate Redis keys
	allTimeKey := fmt.Sprintf("count:%s:all_time", shortURL)
	last24hKey := fmt.Sprintf("count:%s:24h", shortURL)
	lastWeekKey := fmt.Sprintf("count:%s:week", shortURL)

	// Atomic counter update
	err = rdb.Watch(ctx, func(tx *redis.Tx) error {
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Incr(ctx, allTimeKey)
			pipe.Incr(ctx, last24hKey)
			pipe.Expire(ctx, last24hKey, 24*time.Hour) // Set TTL for 24h counter
			pipe.Incr(ctx, lastWeekKey)
			pipe.Expire(ctx, lastWeekKey, 7*24*time.Hour) // Set TTL for week counter
			return nil
		})
		return err
	}, allTimeKey, last24hKey, lastWeekKey)

	if err != nil {
		log.Println("Error updating global counter:", err)
	}
}

// Decrement the global counter ensuring values do not go below zero
func DecrementGlobalCounter(snowflakeID string) error {
	ctx := context.Background()

	// Extract shortURL from the Snowflake ID
	shortURL, err := utils.DecodeShortURLFromSnowflakeID(snowflakeID)
	if err != nil {
		log.Println("Error extracting shortURL from Snowflake ID:", err)
		return err
	}

	// Generate Redis keys
	last24hKey := fmt.Sprintf("count:%s:24h", shortURL)
	lastWeekKey := fmt.Sprintf("count:%s:week", shortURL)

	// Ensure counters do not go below zero
	err = rdb.Watch(ctx, func(tx *redis.Tx) error {
		count24h, _ := tx.Get(ctx, last24hKey).Int()
		countWeek, _ := tx.Get(ctx, lastWeekKey).Int()

		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			if count24h > 0 {
				pipe.Decr(ctx, last24hKey)
			}
			if countWeek > 0 {
				pipe.Decr(ctx, lastWeekKey)
			}
			return nil
		})
		return err
	}, last24hKey, lastWeekKey)

	if err != nil {
		log.Println("Error decrementing global counter:", err)
	}
	return err
}

// Retrieves the count from Redis for a given shortURL with race condition prevention
func GetURLCounter(shortURL string) (int, int, int, error) {
	ctx := context.Background()

	var allTime, last24h, lastWeek int

	err := rdb.Watch(ctx, func(tx *redis.Tx) error {
		pipe := tx.Pipeline()

		allTimeCmd := pipe.Get(ctx, fmt.Sprintf("count:%s:all_time", shortURL))
		last24hCmd := pipe.Get(ctx, fmt.Sprintf("count:%s:24h", shortURL))
		lastWeekCmd := pipe.Get(ctx, fmt.Sprintf("count:%s:week", shortURL))

		_, err := pipe.Exec(ctx)
		if err != nil {
			return err
		}

		allTime, _ = allTimeCmd.Int()
		last24h, _ = last24hCmd.Int()
		lastWeek, _ = lastWeekCmd.Int()

		return nil
	}, fmt.Sprintf("count:%s:all_time", shortURL), fmt.Sprintf("count:%s:24h", shortURL), fmt.Sprintf("count:%s:week", shortURL))

	if err != nil {
		return 0, 0, 0, err
	}

	return allTime, last24h, lastWeek, nil
}
