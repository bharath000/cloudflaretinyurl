package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/mattheath/base62"
)

type URL struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
}

var (
	urlStore    = make(map[string]string)
	lock        = sync.RWMutex{}
	counter     = 14776336
	counterLock = sync.Mutex{}
	baseURL     = "http://localhost:8080/api/v1/"
)

// Generate a unique short URL using a global counter with Base62 encoding
func generateShortURL() string {
	counterLock.Lock()
	counter++
	shortURL := base62.EncodeInt64(int64(counter))
	counterLock.Unlock()
	return shortURL
}

// Create Tiny URL
func createTinyURL(w http.ResponseWriter, r *http.Request) {
	var request URL
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	shortURL := generateShortURL()

	lock.Lock()
	urlStore[shortURL] = request.LongURL
	lock.Unlock()

	response := URL{ShortURL: baseURL + shortURL, LongURL: request.LongURL}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Redirect Tiny URL with 302 status
func redirectTinyURL(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	shortURL := params["shortURL"]

	lock.RLock()
	longURL, exists := urlStore[shortURL]
	lock.RUnlock()

	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound) // 302 Redirect
}

// Delete Tiny URL
func deleteTinyURL(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	shortURL := params["shortURL"]

	lock.Lock()
	_, exists := urlStore[shortURL]
	if exists {
		delete(urlStore, shortURL)
	}
	lock.Unlock()

	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/create", createTinyURL).Methods("POST")
	r.HandleFunc("/api/v1/{shortURL}", redirectTinyURL).Methods("GET")
	r.HandleFunc("/api/v1/delete/{shortURL}", deleteTinyURL).Methods("DELETE")

	http.ListenAndServe(":8080", r)
}
