package redislocks

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func InitRedisLocks(redisClient *redis.Client) {
	rdb = redisClient
}

// Acquire a distributed lock with TTL
func AcquireLock(key string, expiration int) bool {
	return rdb.SetNX(context.Background(), key, "locked", time.Duration(expiration)*time.Second).Val()
}

// Release the distributed lock
func ReleaseLock(key string) {
	rdb.Del(context.Background(), key)
}
