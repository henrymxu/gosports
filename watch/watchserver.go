package watch

import (
	"fmt"
	"github.com/henrymxu/gosportsapi/database"
	"github.com/henrymxu/gosportsapi/sports"
	"github.com/henrymxu/gosportsapi/websocket"
	"github.com/ngaut/log"
	"net/url"
	"strconv"
	"time"
)

const gameChannelStringFormat = "%s%s"

const scheduleCheckDelay = 1 * time.Hour
const gameLiveCheckDelay = 15 * time.Second // TODO: change this to a higher value

type Server struct {
	clientServer *websocket.Server
	databaseServer *database.Server
	gameChannels map[string]*chan websocket.Message
	sports       *sports.Sports
}

func CreateWatchServer(clientServer *websocket.Server, sportsInstance *sports.Sports) *Server {
	server := &Server{
		clientServer: clientServer,
		gameChannels: make(map[string]*chan websocket.Message),
		sports:       sportsInstance,
	}
	go server.watchScheduleForGamesToWatch()
	return server
}

func (s *Server) GetGameChannel(sport sports.Sport, gameId string) *chan websocket.Message {
	return s.gameChannels[fmt.Sprintf(gameChannelStringFormat, sport.Name(), gameId)]
}

func (s *Server) watchScheduleForGamesToWatch() {
	s.parseScheduleForGamesToWatch() // Since ticker starts after the first delay
	ticker := time.NewTicker(scheduleCheckDelay)
	for {
		<-ticker.C
		s.parseScheduleForGamesToWatch()
	}
}

func (s *Server) parseScheduleForGamesToWatch() {
	for _, sportType := range s.sports.RetrieveSportsAsList() {
		sport := sportType
		schedule := sport.Schedule(nil)
		for _, game := range sports.CheckActiveGames(sport, schedule) {
			gameId := game.Id
			gameString := fmt.Sprintf(gameChannelStringFormat, sport.Name(), gameId)
			if _, ok := s.gameChannels[gameString]; !ok { // GameScheduleStatus exists in schedule and does not yet have a channel created
				gameChannel := make(chan websocket.Message)
				s.gameChannels[gameString] = &gameChannel
				s.clientServer.RegisterChannelChannel <- websocket.RegisterChannel{
					Channel: &gameChannel,
					Action:  true,
				}
				delay := game.StartTime.Sub(time.Now()) //TODO: possibly need more grace time to check if game
				go s.waitToWatchGame(&sport, gameId, delay)
			}
			for gameString, gameChannel := range s.gameChannels { // Remove games that are not in the schedule anymore
				if _, ok := s.gameChannels[gameString]; !ok {
					log.Debugf("Closing game for %s with Id %s", sport.Name(), gameId)
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

func (s *Server) waitToWatchGame(sport *sports.Sport, gameId string, delay time.Duration) {
	log.Debugf("Waiting for %s game %s to be live. ETA %s", (*sport).Name(), gameId, delay.String())
	<-time.After(delay)
	go s.watchGame(sport, gameId)
}

type internalGameStatus struct {
	state         sports.ScheduleState
	lastCheck     time.Time
	currentPeriod int
}

func (s *Server) watchGame(sport *sports.Sport, gameId string) {
	ticker := time.NewTicker(gameLiveCheckDelay)
	gameStatus := internalGameStatus {
		state:         sports.Preview,
		lastCheck:     time.Now(),
		currentPeriod: 1,
	}
	for {
		<-ticker.C
		gameStatus = s.parseGame(sport, gameId, gameStatus)
		if gameStatus.state == sports.Complete { // GameScheduleStatus is over, no need to watch
			break
		}
	}
}

func (s *Server) parseGame(sport *sports.Sport, gameId string, prevGameStatus internalGameStatus) internalGameStatus {
	channel := s.gameChannels[fmt.Sprintf(gameChannelStringFormat, (*sport).Name(), gameId)]

	values := url.Values{}
	values.Add("gameId", gameId)
	values.Add("date", sports.CreateDetailedStringFromDate(prevGameStatus.lastCheck))
	values.Add("period", strconv.Itoa(prevGameStatus.currentPeriod))
	playbyplay := (*sport).PlayByPlay(values)
	prevGameStatus.lastCheck = time.Now() // TODO: convert time to string
	if playbyplay == nil { // Should never be nil
		prevGameStatus.state = sports.Preview
		return prevGameStatus
	}
	log.Debugf("Retrieved playbyplay for %s game %s, lastCheck %v", (*sport).Name(), gameId, prevGameStatus.lastCheck.String())
	state, _ :=  playbyplay["metadata"].(map[string]interface{})["state"]
	prevGameStatus.state = sports.ScheduleState(state.(int))
	period, ok  := playbyplay["status"].(map[string]interface{})["period"]
	if ok {
		prevGameStatus.currentPeriod = period.(int)
	}
	message := websocket.Message{
		Type: fmt.Sprintf("playbyplay update at %s", time.Now()),
		Contents: playbyplay,
	}
	*channel<-message
	//s.databaseServer.GetDatabase((*sport).Name()).Collection(strconv.Itoa(gameId)).InsertGameSnapshot(context.Background(), playbyplay)
	return prevGameStatus
}
