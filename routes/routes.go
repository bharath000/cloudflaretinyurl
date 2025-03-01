package routes

import (
	"cloudflaretinyurl/handlers"

	"github.com/gorilla/mux"
)

// Initialize API Routes
func InitRoutes() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/create", handlers.CreateTinyURL).Methods("POST")
	r.HandleFunc("/api/v1/{shortURL}", handlers.RedirectTinyURL).Methods("GET")
	r.HandleFunc("/api/v1/delete/{shortURL}", handlers.DeleteTinyURL).Methods("DELETE")
	r.HandleFunc("/api/v1/clicks/{shortURL}", handlers.GetTinyURLCounts).Methods("GET")
	r.HandleFunc("/api/v1/clicks_fallback/{shortURL}", handlers.GetClickCountsHandler).Methods("GET")
	return r
}
