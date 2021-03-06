package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/henrymxu/gosports/database"
	"github.com/henrymxu/gosports/sports"
	"github.com/henrymxu/gosports/watch"
	"github.com/henrymxu/gosports/websocket"
	"github.com/ngaut/log"
	"net/http"
	"net/url"
)

type server struct {
	stream *watch.Server
	client *websocket.Server
	db     *database.Server
	sports *sports.Sports
	router *mux.Router
}

type httpError struct {
	code int
	text string
}

type WebsocketHandlerFunc func(*websocket.Client, http.ResponseWriter, *http.Request)

type ValidateParameter func(map[string]string) (interface{}, *httpError)

type ValidateQuery func(url.Values) (interface{}, *httpError)

func (s *server) handleAbout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "About")
	}
}

func (s *server) handleSchedule() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		query := r.URL.Query()
		sportInterface, err := parseSport(params)
		if err != nil {
			logHttpError(w, err)
			return
		}
		sport := s.sports.ParseSportId(sportInterface.(int))
		result := sport.Schedule(query)
		games := result["content"].([]map[string]interface{})
		for _, game := range games {
			b, _ := json.MarshalIndent(game, "", "  ")
			_, _ = fmt.Fprintf(w, "ScheduledGame %s:", string(b))
		}
	}
}

func (s *server) handlePlayByPlay() WebsocketHandlerFunc {
	return func(ws *websocket.Client, w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		query := r.URL.Query()
		sportInterface, err := parseSport(params)
		if err != nil {

		}
		sport := s.sports.ParseSportId((sportInterface).(int))
		gameIdInterface, _ := parseGameId(query)

		pbpChannel := s.stream.GetGameChannel(sport, gameIdInterface.(string))
		s.client.RegisterClientToWriteChannel(ws, pbpChannel)
		result := sport.PlayByPlay(query)
		message := websocket.Message {
			Type: "initial playbyplay",
			Contents: result,
		}
		s.client.WriteToClient(ws, message)
	}
}

func parseSport(params map[string]string) (interface{}, *httpError) {
	sportString, ok := params["sport"]
	if !ok {
		return nil, &httpError{
			http.StatusBadRequest,
			"Missing required {sport} parameter",
		}
	}
	sport := sports.ParseSportString(sportString)
	if sport == -1 {
		return nil, &httpError{
			http.StatusBadRequest,
			"Invalid {sport} parameter",
		}
	}
	return sport, nil
}

func parseGameId(query url.Values) (interface{}, *httpError) {
	gameId := query.Get("gameId")
	if gameId == "" {
		return "", &httpError{
			http.StatusBadRequest,
			"Missing required {gameId} query",
		}
	}
	return gameId, nil
}

func (s *server) websocketUpgrade(h WebsocketHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsClient := s.client.BaseWebsocketHandler(w, r)
		if wsClient != nil {
			h(wsClient, w, r)
		}
	}
}

func (s *server) checkValidQueries(h http.HandlerFunc, paramsToValidate []ValidateParameter, queriesToValidate []ValidateQuery) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		query := r.URL.Query()
		for _, validator := range paramsToValidate {
			if _, err := validator(params); err != nil {
				logHttpError(w, err)
				return
			}
		}
		for _, validator := range queriesToValidate {
			if _, err := validator(query); err != nil {
				logHttpError(w, err)
				return
			}
		}
		h(w, r)
	}
}

func logHttpError(w http.ResponseWriter, error *httpError) {
	log.Errorf("Logging http.Error: %s", error.text)
	http.Error(w, error.text, error.code)
}

/*
func (s *server) handleTemplate(files text...) http.HandlerFunc {
	var (
		init sync.Once
		tpl  *template.Template
		err  error
	)
	return func(w http.ResponseWriter, r *http.Request) {
		init.Do(func(){
			tpl, err = template.ParseFiles(files...)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// use tpl
	}
}
*/