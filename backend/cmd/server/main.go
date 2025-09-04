package main

import (
	"log"

	"backend/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Server starting on %s", cfg.GetServerAddress())
	log.Printf("Database DSN: %s", cfg.Database.DSN)
	log.Printf("JWT Issuer: %s", cfg.JWT.Issuer)
	log.Printf("JWT Expiration: %v", cfg.GetJWTExpiration())

	// TODO: Initialize database connection
	// TODO: Initialize Gin router
	// TODO: Setup handlers and middleware
	// TODO: Start server

	log.Println("Tournament backend server setup complete")
}