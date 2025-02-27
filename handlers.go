package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"tinyurl-api/database"

	"github.com/gorilla/mux"
	"github.com/mattheath/base62"
	"github.com/redis/go-redis/v9"
)

type URL struct {
	ShortURL  string     `json:"short_url"`
	LongURL   string     `json:"long_url"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

var baseURL = "http://localhost:8080/api/v1/"

// Generate Unique Short URL
func generateShortURL() string {
	newCounter, err := database.IncrementGlobalCounter()
	if err != nil {
		log.Fatal("Failed to increment global counter:", err)
	}
	return base62.EncodeInt64(newCounter)
}

// Create Short URL Handler
func CreateTinyURL(w http.ResponseWriter, r *http.Request) {
	var request URL
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	shortURL := generateShortURL()

	// Store in PostgreSQL
	err := database.StoreURL(shortURL, request.LongURL)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Cache in Redis
	database.CacheURL(shortURL, request.LongURL)

	response := URL{ShortURL: baseURL + shortURL, LongURL: request.LongURL}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Redirect to Original URL
func RedirectTinyURL(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	shortURL := params["shortURL"]

	// Check Redis Cache First
	longURL, err := database.GetCachedURL(shortURL)
	if err == redis.Nil {
		// Fetch from PostgreSQL
		longURL, err = database.GetURL(shortURL)
		if err != nil {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		// Cache the result in Redis
		database.CacheURL(shortURL, longURL)
	}

	// Increment Access Count
	_ = database.RDB.Incr(context.Background(), fmt.Sprintf("count:%s:all_time", shortURL))
	_ = database.RDB.Incr(context.Background(), fmt.Sprintf("count:%s:24h", shortURL))
	_ = database.RDB.Incr(context.Background(), fmt.Sprintf("count:%s:week", shortURL))

	// Store Click Timestamp in PostgreSQL
	_, err = database.DB.Exec("INSERT INTO url_clicks (short_url, accessed_at) VALUES ($1, NOW())", shortURL)
	if err != nil {
		log.Println("Failed to log click event:", err)
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}

// Delete Short URL Handler
func DeleteTinyURL(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	shortURL := params["shortURL"]

	// Delete from PostgreSQL
	_, err := database.DB.Exec("DELETE FROM urls WHERE short_url=$1", shortURL)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Remove from Redis
	database.RDB.Del(context.Background(), shortURL)
	database.RDB.Del(context.Background(), fmt.Sprintf("count:%s:all_time", shortURL))
	database.RDB.Del(context.Background(), fmt.Sprintf("count:%s:24h", shortURL))
	database.RDB.Del(context.Background(), fmt.Sprintf("count:%s:week", shortURL))

	w.WriteHeader(http.StatusNoContent)
}
