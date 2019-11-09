package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/henrymxu/gosportsapi/database"
	"github.com/henrymxu/gosportsapi/sports"
	"github.com/henrymxu/gosportsapi/stream"
	"github.com/henrymxu/gosportsapi/websocket"
	"github.com/ngaut/log"
	"net/http"
	"net/url"
	"strconv"
)

type server struct {
	stream *stream.Server
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
		fmt.Fprint(w, "About")
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
		sport := s.sports.ParseSportId(convertInterfaceToInt(sportInterface))
		result := sport.Schedule(query)
		games := result["content"].([]map[string]interface{})
		for _, game := range games {
			b, _ := json.MarshalIndent(game, "", "  ")
			_, _ = fmt.Fprintf(w, "Game %s:", string(b))
		}
	}
}

func (s *server) handlePlayByPlay() WebsocketHandlerFunc {
	return func(ws *websocket.Client, w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		query := r.URL.Query()
		sportInterface, _ := parseSport(params)
		sport := s.sports.ParseSportId(convertInterfaceToInt(sportInterface))
		gameIdInterface, _ := parseGameId(query)
		gameId := convertInterfaceToInt(gameIdInterface)

		pbpChannel := s.stream.GetGameChannel(sport, gameId)
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

func convertInterfaceToInt(iface interface{}) int {
	return iface.(int)
}

func parseGameId(query url.Values) (interface{}, *httpError) {
	gameId, err := strconv.Atoi(query.Get("gameId"))
	if err != nil {
		return 0, &httpError{
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

func (s *server) handleGreeting(format text) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, format, "World")
	}
}

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