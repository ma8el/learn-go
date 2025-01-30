package main

import (
	"log"
	"net/http"

	"notes-api/db"
	"notes-api/routes"
)

func main() {
	// Initialize database
	if err := db.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Setup routes
	r := routes.SetupRouter()

	log.Println("Server is running on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}
