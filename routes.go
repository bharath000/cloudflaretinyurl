package routes

import (
	"tinyurl-api/handlers"

	"github.com/gorilla/mux"
)

// Initialize API Routes
func InitRoutes() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/create", handlers.CreateTinyURL).Methods("POST")
	r.HandleFunc("/api/v1/{shortURL}", handlers.RedirectTinyURL).Methods("GET")
	r.HandleFunc("/api/v1/delete/{shortURL}", handlers.DeleteTinyURL).Methods("DELETE")
	return r
}
