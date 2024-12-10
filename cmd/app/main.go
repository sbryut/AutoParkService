package main

import (
	"log"
	"net/http"

	"AutoParkWeb/internal/config"
	"AutoParkWeb/internal/database/postgres"
	"AutoParkWeb/internal/services"
	"AutoParkWeb/internal/transport"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	service := services.NewAutoParkService(db)
	router := transport.SetupRoutes(service)

	log.Println("Server started on 127.0.0.1:8080")
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
