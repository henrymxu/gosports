package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/henrymxu/gosportsapi/client"
	"github.com/henrymxu/gosportsapi/database"
	"github.com/henrymxu/gosportsapi/sports"
	"github.com/ngaut/log"
	"net/http"
)

type server struct {
	client *client.Server
	db     *database.Server
	sports *sports.Sports
	router *mux.Router
}

type httpError struct {
	code int
	string string
}

type WebsocketHandlerFunc func(*client.WebsocketClient, http.ResponseWriter, *http.Request)

func (s *server) handleAbout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "About")
	}
}

func (s *server) handleSchedule() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		query := r.URL.Query()
		sport, httpError := s.parseSport(params)
		if httpError != nil {
			http.Error(w, httpError.string, httpError.code)
			return
		}
		result := sport.Schedule(query)
		games := result["content"].([]map[string]interface{})
		for _, game := range games {
			b, _ := json.MarshalIndent(game, "", "  ")
			_, _ = fmt.Fprintf(w, "Game %s:", string(b))
		}
	}
}

func (s *server) handlePlayByPlay() WebsocketHandlerFunc {
	return func(ws *client.WebsocketClient, w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		query := r.URL.Query()
		sport, httpError := s.parseSport(params)
		if httpError != nil {
			log.Errorf("Logging http.Error: %s", httpError.string)
			http.Error(w, httpError.string, httpError.code)
			return
		}
		gameId := query.Get("gameId")
		if gameId == "" {
			//Missing gameId
			log.Errorf("Logging http.Error: %s", http.StatusText(http.StatusUnprocessableEntity))
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		}
		// get channel from sport with gameId
		pbpChannel := make(chan client.Message)
		// channel should already be registered
		s.client.RegisterChannelChannel <- &pbpChannel
		// register client to channel
		for !s.client.RegisterClientToWriteChannel(ws, &pbpChannel) {

		}
		// maybe ask to populate previous plays?
		message := client.Message{
			Contents: map[string]interface{}{"sport": sport, "sportId": params["sport"], "gameId": gameId},
		}
		pbpChannel <- message
	}
}

func (s *server) parseSport(params map[string]string) (sports.Sport, *httpError) {
	sportString, ok := params["sport"]
	if !ok {
		return nil, &httpError{
			http.StatusBadRequest,
			"Missing required {sport} parameter",
		}
	}
	sport := s.sports.ParseSportString(sportString)
	if sport == nil {
		return nil, &httpError{
			http.StatusUnprocessableEntity,
			http.StatusText(http.StatusUnprocessableEntity),
		}
	}
	return sport, nil
}

func (s *server) websocketUpgrade(h WebsocketHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsClient := s.client.BaseWebsocketHandler(w, r)
		if wsClient != nil {
			h(wsClient, w, r)
		}
	}
}

/*

func (s *server) handleGreeting(format string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, format, "World")
	}
}

func (s *server) handleTemplate(files string...) http.HandlerFunc {
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