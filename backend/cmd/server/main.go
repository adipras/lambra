package main

import (
	"fmt"
	"log"

	"github.com/yourusername/lambra/internal/api/router"
	"github.com/yourusername/lambra/internal/config"
	"github.com/yourusername/lambra/internal/database"
)

func main() {
	log.Println("Starting Lambra Service Generator...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connection established")

	// Setup router
	r := router.Setup(db)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
