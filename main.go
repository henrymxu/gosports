package main

import (
	"github.com/gorilla/mux"
	"github.com/henrymxu/gosports/database"
	"github.com/henrymxu/gosports/sports"
	"github.com/henrymxu/gosports/watch"
	"github.com/henrymxu/gosports/websocket"
	"log"
	"net/http"
	"time"
)

const serverAddress = "localhost:8080"
const databaseAddress = "mongodb://localhost:27017"

func main() {
	databaseClient := database.MongoClient{}
	databaseClient.Initialize(databaseAddress)
	databaseServer := database.CreateDatabaseServer(&databaseClient)

	websocketServer := websocket.CreateWebsocketServer()

	sportsInstance := sports.InitializeSports()
	streamServer := watch.CreateWatchServer(websocketServer, sportsInstance)

	router := mux.NewRouter()
	server := server{
		stream: streamServer,
		client: websocketServer,
		db:     databaseServer,
		sports: sportsInstance,
		router: router,
	}

	server.routes()

	srv := &http.Server{
		Handler:      server.router,
		Addr:         serverAddress,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
