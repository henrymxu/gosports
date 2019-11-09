package stream

import (
	"fmt"
	"github.com/henrymxu/gosportsapi/sports"
	"github.com/henrymxu/gosportsapi/websocket"
	"github.com/ngaut/log"
)

const gameChannelStringFormat = "%s%d"

type Server struct {
	clientServer *websocket.Server
	gameChannels map[string]*chan websocket.Message
	sports       *sports.Sports
	tickChannel  <-chan bool
}

func CreateStreamServer(clientServer *websocket.Server, sportsInstance *sports.Sports, tick <-chan bool) *Server {
	server := &Server{
		clientServer: clientServer,
		gameChannels: make(map[string]*chan websocket.Message),
		sports:       sportsInstance,
		tickChannel:  tick,
	}
	go server.watchScheduleForGamesToStream()
	return server
}

func (s *Server) GetGameChannel(sport sports.Sport, gameId int) *chan websocket.Message {
	return s.gameChannels[fmt.Sprintf(gameChannelStringFormat, sport.Name(), gameId)]
}

func (s *Server) watchScheduleForGamesToStream() {
	for {
		<-s.tickChannel
		for _, sport := range s.sports.RetrieveSportsAsList() {
			name := sport.Name()
			schedule := sport.Schedule(nil)
			for _, game := range sport.CheckActiveGames(schedule) {
				gameId := game.Id
				gameString := fmt.Sprintf(gameChannelStringFormat, name, gameId)
				if _, ok := s.gameChannels[gameString]; !ok { // Game exists in schedule and does not yet have a channel created
					log.Debugf("Registering new game for %s with Id %d", name, gameId)
					gameChannel := make(chan websocket.Message)
					s.gameChannels[gameString] = &gameChannel
					s.clientServer.RegisterChannelChannel <- websocket.RegisterChannel{
						Channel: &gameChannel,
						Action:  true,
					}
				}
				for gameString, gameChannel := range s.gameChannels { // Remove games that are not in the schedule anymore
					if _, ok := s.gameChannels[gameString]; !ok {
						log.Debugf("Closing game for %s with Id %d", name, gameId)
						s.clientServer.RegisterChannelChannel <- websocket.RegisterChannel{
							Channel: gameChannel,
							Action:  false,
						}
						delete(s.gameChannels, gameString)
					}
				}
			}
		}
	}
}
