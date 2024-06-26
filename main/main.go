package main

import (
	"apiGo/api"
	"apiGo/storage"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	// Initialize and start the database.
	db, err := storage.NewPgStorage()
	if err != nil {
		slog.Error("db couldn't start")
		os.Exit(1)
	}

	if err = db.Init(); err != nil {
		slog.Error("db couldn't be initialized")
		os.Exit(1)
	}

	// Create a new instance of the API server.
	apiServer := api.NewApiServer(":8080", db)

	// Set up API endpoints and their handlers.
	apiServer.HandleEndpoints()

	// Start the API server.
	fmt.Println("Server running in 8080...")
	if err := apiServer.Run(); err != nil {
		slog.Error("server couldn't start")
		os.Exit(1)
	}
}
