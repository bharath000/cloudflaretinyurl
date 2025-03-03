package database

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

var (
	DB  *sql.DB
	RDB *redis.Client
)

// Initialize Database Connections
func InitDB() error {
	postgresURL := os.Getenv("DATABASE_URL")
	if postgresURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	log.Println("Connecting to PostgreSQL at:", postgresURL)

	var err error
	DB, err = sql.Open("postgres", postgresURL)
	if err != nil {
		log.Fatalf("Failed to open PostgreSQL connection: %v", err)
	}

	// Verify Database Connection
	// err = DB.Ping()
	// if err != nil {
	// 	log.Fatalf("Failed to ping PostgreSQL: %v", err)
	// }
	// log.Println("✅ Connected to PostgreSQL successfully!")

	// Initialize Redis
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "cloudflaretinyurl_redis:6379"
	}

	RDB = redis.NewClient(&redis.Options{Addr: redisURL})
	if _, err := RDB.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis successfully!")

	return nil
}

// Generate Global Counter for Unique Short URLs
func IncrementGlobalCounter() (int64, error) {
	return RDB.Incr(context.Background(), "url_global_counter").Result()
}

func StoreURL(shortURL, longURL string, expiresAt *time.Time) error {
	_, err := DB.Exec("INSERT INTO urls (short_url, long_url, created_at, expires_at) VALUES ($1, $2, NOW(), $3)",
		shortURL, longURL, expiresAt)
	if err != nil {
		log.Printf("Database insertion error: %v", err)
	}
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

// Get click counts from PostgreSQL
func GetClickCounts(shortURL string) (int, int, int, error) {
	var allTime, last24h, lastWeek int

	// Get all-time clicks
	err := DB.QueryRow("SELECT COUNT(*) FROM url_clicks WHERE short_url=$1", shortURL).Scan(&allTime)
	if err != nil {
		return 0, 0, 0, err
	}

	// Get last 24 hours clicks
	err = DB.QueryRow("SELECT COUNT(*) FROM url_clicks WHERE short_url=$1 AND accessed_at >= NOW() - INTERVAL '24 hours'", shortURL).Scan(&last24h)
	if err != nil {
		return 0, 0, 0, err
	}

	// Get last week clicks
	err = DB.QueryRow("SELECT COUNT(*) FROM url_clicks WHERE short_url=$1 AND accessed_at >= NOW() - INTERVAL '7 days'", shortURL).Scan(&lastWeek)
	if err != nil {
		return 0, 0, 0, err
	}

	return allTime, last24h, lastWeek, nil
}

// GetShortURLByLongURL checks if a long URL already exists and returns its short URL & expiry date
func GetShortURLByLongURL(longURL string) (string, *time.Time, error) {
	var shortURL string
	var expiresAt sql.NullTime

	err := DB.QueryRow("SELECT short_url, expires_at FROM urls WHERE long_url = $1", longURL).
		Scan(&shortURL, &expiresAt)

	if err != nil {
		return "", nil, err
	}

	// Convert sql.NullTime to *time.Time
	if expiresAt.Valid {
		return shortURL, &expiresAt.Time, nil
	}

	return shortURL, nil, nil
}
