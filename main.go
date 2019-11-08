package main

import (
	"github.com/gorilla/mux"
	"github.com/henrymxu/gosportsapi/client"
	"github.com/henrymxu/gosportsapi/database"
	"github.com/henrymxu/gosportsapi/sports"
	"log"
	"net/http"
	"time"
)

func main() {
	databaseClient := database.MongoClient{}
	databaseClient.Initialize("mongodb://localhost:27017")
	websocketServer := client.CreateClientServer()
	sportsInstance := sports.InitializeSports()

	server := server{
		websocketServer,
		database.CreateDatabaseServer(&databaseClient),
		sportsInstance,
		mux.NewRouter(),
	}

	server.routes()

	srv := &http.Server{
		Handler:      server.router,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
