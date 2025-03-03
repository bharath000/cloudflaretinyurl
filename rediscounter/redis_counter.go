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
	last1minKey := fmt.Sprintf("count:%s:1min", shortURL)

	// Atomic counter update
	err = rdb.Watch(ctx, func(tx *redis.Tx) error {
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Incr(ctx, allTimeKey)
			pipe.Incr(ctx, last1minKey)
			pipe.Expire(ctx, last1minKey, 1*time.Minute) // Set TTL for 1min counter
			pipe.Incr(ctx, last24hKey)
			pipe.Expire(ctx, last24hKey, 24*time.Hour) // Set TTL for 24h counter
			pipe.Incr(ctx, lastWeekKey)
			pipe.Expire(ctx, lastWeekKey, 7*24*time.Hour) // Set TTL for week counter
			return nil
		})
		return err
	}, allTimeKey, last24hKey, lastWeekKey, last1minKey)

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
	last1minKey := fmt.Sprintf("count:%s:1min", shortURL)

	// Ensure counters do not go below zero
	err = rdb.Watch(ctx, func(tx *redis.Tx) error {
		count24h, _ := tx.Get(ctx, last24hKey).Int()
		countWeek, _ := tx.Get(ctx, lastWeekKey).Int()
		count1min, _ := tx.Get(ctx, last1minKey).Int()

		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			if count24h > 0 {
				pipe.Decr(ctx, last24hKey)
			}
			if countWeek > 0 {
				pipe.Decr(ctx, lastWeekKey)
			}
			if count1min > 0 {
				pipe.Decr(ctx, last1minKey)
			}
			return nil
		})
		return err
	}, last24hKey, lastWeekKey, last1minKey)

	if err != nil {
		log.Println("Error decrementing global counter:", err)
	}
	return err
}

// Retrieves the count from Redis for a given shortURL with race condition prevention
func GetURLCounter(shortURL string) (int, int, int, int, error) {
	ctx := context.Background()

	// Define Redis keys
	allTimeKey := fmt.Sprintf("count:%s:all_time", shortURL)
	last1minKey := fmt.Sprintf("count:%s:1min", shortURL)
	last24hKey := fmt.Sprintf("count:%s:24h", shortURL)
	lastWeekKey := fmt.Sprintf("count:%s:week", shortURL)

	// Start a pipeline to fetch multiple keys in one go
	pipe := rdb.Pipeline()

	allTimeCmd := pipe.Get(ctx, allTimeKey)
	last1minCmd := pipe.Get(ctx, last1minKey)
	last24hCmd := pipe.Get(ctx, last24hKey)
	lastWeekCmd := pipe.Get(ctx, lastWeekKey)

	// Execute the pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		log.Println("Error executing Redis pipeline:", err)
		return 0, 0, 0, 0, err
	}

	// Convert Redis responses, ensuring missing values default to 0
	allTime, _ := safeRedisGet(allTimeCmd)
	last1min, _ := safeRedisGet(last1minCmd)
	last24h, _ := safeRedisGet(last24hCmd)
	lastWeek, _ := safeRedisGet(lastWeekCmd)

	return allTime, last24h, lastWeek, last1min, nil
}

// Helper function to safely parse Redis responses, returning 0 for missing keys
func safeRedisGet(cmd *redis.StringCmd) (int, error) {
	val, err := cmd.Int()
	if err == redis.Nil {
		return 0, nil // ✅ Return zero if key is missing or expired
	}
	return val, err
}
