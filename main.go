package main

import (
	"log"
	"net/http"

	"cloudflaretinyurl/database"
	"cloudflaretinyurl/rediscounter"
	"cloudflaretinyurl/redislocks"
	"cloudflaretinyurl/redispubsub"
	"cloudflaretinyurl/redisqueue"
	"cloudflaretinyurl/routes"
	"cloudflaretinyurl/utils"
)

func main() {
	// Initialize PostgreSQL & Redis
	if err := database.InitDB(); err != nil {
		log.Fatalf("Initialization Error: %v", err)
	}

	// Initialize Snowflake ID generator with machine ID 1
	if err := utils.InitSnowflake(1); err != nil {
		log.Fatalf("Failed to initialize Snowflake ID generator: %v", err)
	}

	// Initialize Redis-based services (Counters, Queues, Pub/Sub, Locks)
	rediscounter.InitRedisCounter(database.RDB)
	redisqueue.InitRedisQueue(database.RDB)
	redispubsub.InitRedisPubSub(database.RDB)
	redislocks.InitRedisLocks(database.RDB)

	// Start processing expired clicks in a separate goroutine
	go redisqueue.ProcessExpiredClicks()

	// Start listening for Redis Pub/Sub events
	go redispubsub.ListenForExpiredClicks()

	// Set up API routes
	r := routes.InitRoutes()

	log.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", r)
}
