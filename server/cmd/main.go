package main

import (
	"log"
	"os"
	"server/db"
	"server/internal/user"
	"server/internal/ws"
	"server/router"
)

func main() {
	// Get the PORT from environment variable (set by Heroku)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port for local testing
		log.Printf("PORT not set, using default port %s", port)
	}

	// Initialize database connection
	dbConn, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("Could not initialize database connection: %s", err)
	}

	// Set up user repository, service, and handler
	userRep := user.NewRepository(dbConn.GetDB())
	userSvc := user.NewService(userRep)
	userHandler := user.NewHandler(userSvc)

	// Set up WebSocket hub and handler
	hub := ws.NewHub()
	err = ws.LoadChatsIntoHub(hub, dbConn.GetDB())
	if err != nil {
		log.Fatalf("Failed to load chats into Hub: %v", err)
	}
	go hub.Run(dbConn.GetDB())

	wsHandler := ws.NewHandler(hub, dbConn.GetDB())

	// Initialize and start the router
	router.InitRouter(userHandler, wsHandler)

	// Start the server
	log.Printf("Server is running on port %s", port)
	err = router.Start(":" + port)
	if err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}
}
