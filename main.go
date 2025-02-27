package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"context"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/mattheath/base62"
	"github.com/redis/go-redis/v9"
)

type URL struct {
	ShortURL  string     `json:"short_url"`
	LongURL   string     `json:"long_url"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

var (
	lock    = sync.RWMutex{}
	baseURL = "http://localhost:8080/api/v1/"
	db      *sql.DB
	rdb     *redis.Client
)

func init() {
	var err error
	db, err = sql.Open("postgres", "user=youruser password=yourpassword dbname=tinyurl sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func generateShortURL() string {
	// Increment global counter in Redis to ensure unique values
	newCounter, err := rdb.Incr(context.Background(), "url_global_counter").Result()
	if err != nil {
		log.Fatal("Failed to increment global counter in Redis:", err)
	}

	// Convert counter to Base62 to generate short URL
	shortURL := base62.EncodeInt64(newCounter)

	return shortURL
}

func createTinyURL(w http.ResponseWriter, r *http.Request) {
	var request URL
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	shortURL := generateShortURL()
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	// Store in PostgreSQL
	_, err := db.Exec("INSERT INTO urls (short_url, long_url, created_at, expires_at) VALUES ($1, $2, NOW(), $3)", shortURL, request.LongURL, expiresAt)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Cache in Redis
	rdb.Set(context.Background(), shortURL, request.LongURL, 24*time.Hour)

	response := URL{ShortURL: baseURL + shortURL, LongURL: request.LongURL}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func redirectTinyURL(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	shortURL := params["shortURL"]

	longURL, err := rdb.Get(context.Background(), shortURL).Result()
	if err == redis.Nil {
		lock.RLock()
		err = db.QueryRow("SELECT long_url FROM urls WHERE short_url=$1", shortURL).Scan(&longURL)
		lock.RUnlock()
		if err != nil {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		// Cache the result in Redis
		rdb.Set(context.Background(), shortURL, longURL, 24*time.Hour)
	}

	// Increment access count
	_ = rdb.Incr(context.Background(), fmt.Sprintf("count:%s:all_time", shortURL))
	_ = rdb.Incr(context.Background(), fmt.Sprintf("count:%s:24h", shortURL))
	_ = rdb.Incr(context.Background(), fmt.Sprintf("count:%s:week", shortURL))

	// Store Click Timestamp in PostgreSQL
	_, err = db.Exec("INSERT INTO url_clicks (short_url, accessed_at) VALUES ($1, NOW())", shortURL)
	if err != nil {
		log.Println("Failed to log click event:", err)
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}

func deleteTinyURL(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	shortURL := params["shortURL"]

	lock.Lock()
	_, err := db.Exec("DELETE FROM urls WHERE short_url=$1", shortURL)
	lock.Unlock()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	rdb.Del(context.Background(), shortURL)
	rdb.Del(context.Background(), fmt.Sprintf("count:%s:all_time", shortURL))
	rdb.Del(context.Background(), fmt.Sprintf("count:%s:24h", shortURL))
	rdb.Del(context.Background(), fmt.Sprintf("count:%s:week", shortURL))

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/create", createTinyURL).Methods("POST")
	r.HandleFunc("/api/v1/{shortURL}", redirectTinyURL).Methods("GET")
	r.HandleFunc("/api/v1/delete/{shortURL}", deleteTinyURL).Methods("DELETE")

	http.ListenAndServe(":8080", r)
}
