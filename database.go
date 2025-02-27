package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	DB  *sql.DB
	RDB *redis.Client
)

// Initialize Database Connections
func InitDB() error {
	var err error
	DB, err = sql.Open("postgres", "user=youruser password=yourpassword dbname=tinyurl sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	RDB = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	// Check if Redis is reachable
	if _, err := RDB.Ping(context.Background()).Result(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return nil
}

// Generate Global Counter for Unique Short URLs
func IncrementGlobalCounter() (int64, error) {
	return RDB.Incr(context.Background(), "url_global_counter").Result()
}

// Store URL in PostgreSQL
func StoreURL(shortURL, longURL string) error {
	_, err := DB.Exec("INSERT INTO urls (short_url, long_url, created_at, expires_at) VALUES ($1, $2, NOW(), $3)",
		shortURL, longURL, time.Now().Add(30*24*time.Hour))
	return err
}

// Fetch URL from PostgreSQL
func GetURL(shortURL string) (string, error) {
	var longURL string
	err := DB.QueryRow("SELECT long_url FROM urls WHERE short_url=$1", shortURL).Scan(&longURL)
	return longURL, err
}

// Cache URL in Redis
func CacheURL(shortURL, longURL string) {
	RDB.Set(context.Background(), shortURL, longURL, 24*time.Hour)
}

// Fetch URL from Redis
func GetCachedURL(shortURL string) (string, error) {
	return RDB.Get(context.Background(), shortURL).Result()
}
