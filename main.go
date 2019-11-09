package main

import (
	"github.com/gorilla/mux"
	"github.com/henrymxu/gosportsapi/websocket"
	"github.com/henrymxu/gosportsapi/database"
	"github.com/henrymxu/gosportsapi/sports"
	"github.com/henrymxu/gosportsapi/stream"
	"log"
	"net/http"
	"time"
)

func main() {
	databaseClient := database.MongoClient{}
	databaseClient.Initialize("mongodb://localhost:27017")
	websocketServer := websocket.CreateWebsocketServer()
	sportsInstance := sports.InitializeSports()

	databaseServer := database.CreateDatabaseServer(&databaseClient)
	streamServer := stream.CreateStreamServer(websocketServer, sportsInstance, databaseServer.GetDatabaseTickChannel())

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
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
