package watch

import (
	"fmt"
	"github.com/henrymxu/gosports/database"
	"github.com/henrymxu/gosports/sports"
	"github.com/henrymxu/gosports/websocket"
	"github.com/ngaut/log"
	"net/url"
	"strconv"
	"time"
)

const gameChannelStringFormat = "%s%s"

const scheduleCheckDelay = 1 * time.Hour
const gameLiveCheckDelay = 20 * time.Second // TODO: change this to a higher value

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
	for _, sportType := range *s.sports {
		sport := sportType
		schedule := sport.Schedule(nil)
		for _, game := range sports.CheckActiveGames(sport, schedule) {
			gameId := game.Id
			gameString := fmt.Sprintf(gameChannelStringFormat, sport.Name(), gameId)
			if _, ok := s.gameChannels[gameString]; !ok { // ScheduledGame exists in schedule and does not yet have a channel created
				gameChannel := make(chan websocket.Message)
				s.gameChannels[gameString] = &gameChannel
				s.clientServer.RegisterChannelChannel <- websocket.RegisterChannel{
					Channel: &gameChannel,
					Action:  true,
				}
				go s.waitToWatchGame(&sport, game)
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

func (s *Server) waitToWatchGame(sport *sports.Sport, game sports.ScheduledGame) {
	delay := game.StartTime.Sub(time.Now()) //TODO: possibly need more grace time to check if game
	home := game.Details["home"].(map[string]interface{})["name"]
	away := game.Details["away"].(map[string]interface{})["name"]
	log.Debugf("Waiting for %s vs %s (%s: %s) to be live. ETA %s", home, away, (*sport).Name(), game.Id, delay.String())
	<-time.After(delay)
	go s.watchGame(sport, game)
}

type internalGameStatus struct {
	state         sports.ScheduleState
	lastCheck     string
	currentPeriod int
}

func (s *Server) watchGame(sport *sports.Sport, game sports.ScheduledGame) {
	ticker := time.NewTicker(gameLiveCheckDelay)
	gameStatus := internalGameStatus {
		state:         sports.Preview,
		lastCheck:     (*sport).DefaultTimeString(),
		currentPeriod: 1,
	}
	for {
		<-ticker.C
		gameStatus = s.parseGame(sport, game, gameStatus)
		if gameStatus.state == sports.Complete { // ScheduledGame is over, no need to watch
			log.Debugf("Game complete (%s: %s)", (*sport).Name(), game.Id)
			break
		}
	}
}

func (s *Server) parseGame(sport *sports.Sport, game sports.ScheduledGame, prevGameStatus internalGameStatus) internalGameStatus {
	channel := s.gameChannels[fmt.Sprintf(gameChannelStringFormat, (*sport).Name(), game.Id)]
	values := url.Values{}
	values.Add("gameId", game.Id)
	values.Add("date", prevGameStatus.lastCheck)
	values.Add("period", strconv.Itoa(prevGameStatus.currentPeriod))
	playbyplay := (*sport).PlayByPlay(values)
	if playbyplay == nil { // Should never be nil
		prevGameStatus.state = sports.Preview
		return prevGameStatus
	}
	// Debugging information
	home := game.Details["home"].(map[string]interface{})["name"]
	away := game.Details["away"].(map[string]interface{})["name"]
	log.Debugf("%s vs %s (%s: %s), lastCheck %v", home, away, (*sport).Name(), game.Id, prevGameStatus.lastCheck)
	prevGameStatus.lastCheck = playbyplay["metadata"].(map[string]interface{})["lastCheck"].(string)
	state, _ := playbyplay["metadata"].(map[string]interface{})["state"]
	prevGameStatus.state = state.(sports.ScheduleState)
	period, ok := playbyplay["game"].(map[string]interface{})["status"].(map[string]interface{})["period"]
	if ok {
		prevGameStatus.currentPeriod = period.(int)
	}
	//log.Debugf("Length of plays: %d", len(playbyplay["plays"].([]map[string]interface{})))
	var message websocket.Message
	if prevGameStatus.state != sports.Intermission {
		message = websocket.Message{
			Type:     fmt.Sprintf("playbyplay update at %s", time.Now()),
			Contents: playbyplay,
		}
	} else {
		message = websocket.Message{
			Type:     fmt.Sprintf("playbyplay update at %s", time.Now()),
			Contents: map[string]interface{}{"contents": "intermission"},
		}
	}
	*channel<-message
	//s.databaseServer.GetDatabase((*sport).Name()).Collection(strconv.Itoa(gameId)).InsertGameSnapshot(context.Background(), playbyplay)
	return prevGameStatus
}
