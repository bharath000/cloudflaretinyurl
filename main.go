package main

import (
	database "cloudflaretinyurl"
	"log"
	"net/http"

	"tinyurl-api/database"
	"tinyurl-api/routes"
)

func main() {
	// Initialize PostgreSQL & Redis
	if err := database.InitDB(); err != nil {
		log.Fatalf("Initialization Error: %v", err)
	}

	// Set up API routes
	r := routes.InitRoutes()

	log.Println("Server is running on port 8080...")
	http.ListenAndServe(":8080", r)
}
