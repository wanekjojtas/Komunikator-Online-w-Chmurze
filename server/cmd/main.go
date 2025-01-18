package main

import (
	"log"
	"server/db"
	"server/internal/user"
	"server/internal/ws"
	"server/router"
)

func main() {
	dbConn, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("could not initialize database connection: %s", err)
	}

	userRep := user.NewRepository(dbConn.GetDB())
	userSvc := user.NewService(userRep)
	userHandler := user.NewHandler(userSvc)

	hub := ws.NewHub()
	err = ws.LoadChatsIntoHub(hub, dbConn.GetDB())
	if err != nil {
    log.Fatalf("Failed to load chats into Hub: %v", err)
	}
	go hub.Run()

	wsHandler := ws.NewHandler(hub, dbConn.GetDB())

	router.InitRouter(userHandler, wsHandler)
	router.Start("0.0.0.0:8080")
}
