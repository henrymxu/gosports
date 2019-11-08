package database

import (
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/ngaut/log"
	"golang.org/x/net/context"
	"time"
)

const FormatDatabaseName = "%sGameData" //Example NHLGameData or NBAGameData

type Server struct {
	Client              client // TODO convert this to a interface?
	newLiveChannel      chan int
	completedChannel    chan int
	currentLiveChannels map[int]bool
	tickChannels        map[chan bool]bool // This needs concurrency support
}

func CreateDatabaseServer(client client) *Server {
	server := Server{
		Client:            client,
		tickChannels:        make(map[chan bool]bool),
		newLiveChannel:      make(chan int),
		completedChannel:    make(chan int),
		currentLiveChannels: make(map[int]bool),
	}
	go server.initializeTicker(5)
	return &server
}

// Retrieve the database corresponding to a sport
func (d *Server) GetDatabase(sport string) Database {
	return d.Client.Database(fmt.Sprintf(FormatDatabaseName, sport))
}

// Creates a new tick channel that replicates the database tick channel
func (d *Server) GetDatabaseTickChannel() <-chan bool {
	channel := make(chan bool)
	d.tickChannels[channel] = true
	return channel
}

// Returns a channel for that watches a collection for new documents
func (d *Server) WatchCollection(collection Collection) <-chan bson.Raw {
	log.Debugf("Registering a new collection watcher")
	watchChannel := make(chan bson.Raw)
	go func() {
		cursor, err := collection.WatchGame(context.Background())
		if err != nil {
			log.Error(err.Error())
			return
		}
		defer cursor.Close()
		for cursor.Next() {
			bsonRaw := cursor.Decode()
			log.Debugf("Sending value to watchChannel %s", bsonRaw.String())
			watchChannel <- bsonRaw
		}
	}()
	return watchChannel
}

// initializeTicker creates a channel that is sent to every rate seconds
func (d *Server) initializeTicker(rate time.Duration) {
	log.Debugf("Initializing DatabaseTicker with rate %v", rate)
	if rate == 0 {
		rate = 5
	}
	ticker := time.NewTicker(rate * time.Second)
	for {
		<-ticker.C
		for channel := range d.tickChannels {
			channel <- true
		}
	}
}
