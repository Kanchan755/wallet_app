package main

import (
	"log"

	"github.com/kanchan755/wallet_app/monolith/internal/config"
	"github.com/kanchan755/wallet_app/monolith/internal/database"
)

func main() {
	log.Println("Starting the monolith application...")
	// Load configuration
	cfg := config.Loadconfig()

	// Connect to the database
	db, err := database.ConnectWithRetry(cfg.DBDSN, 5)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer db.Close()
	log.Println("Application Succesfully initilaized ")
}
