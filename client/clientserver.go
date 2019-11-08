package client

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ngaut/log"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	ClientMessageChannel   chan Message // Used for messages received from clients
	quitChannel            chan *WebsocketClient
	clients                map[*WebsocketClient]bool
	RegisterChannelChannel chan *chan Message                   // Used to register new available channels
	writeMessageChannels   map[*chan Message][]*WebsocketClient //TODO: threadsafe
}

type WebsocketClient struct {
	Socket  *websocket.Conn // Websocket connection
	Channel *chan Message   // TODO: not used (Channel the websocket is registered to (Only 1 Channel is supported))
}

// Define our message object
type Message struct {
	Client   *WebsocketClient // Where the message came from
	Id       int
	Contents map[string]interface{}
}

var upgrader = websocket.Upgrader{} // websocket upgrader

func CreateClientServer() *Server {
	server := Server{
		ClientMessageChannel:   make(chan Message),
		quitChannel:            make(chan *WebsocketClient),
		clients:                make(map[*WebsocketClient]bool),
		RegisterChannelChannel: make(chan *chan Message),
		writeMessageChannels:   make(map[*chan Message][]*WebsocketClient),
	}

	go server.closeWebsocket()
	go server.registerNewChannels()

	go func() {
		for {
			message := <-server.ClientMessageChannel
			log.Debugf("Routing message from %s", message.Client.Socket.RemoteAddr())
		}
	}()

	return &server
}

func (s *Server) BaseWebsocketHandler(w http.ResponseWriter, r *http.Request) *WebsocketClient {
	requestOrigin := r.Header.Get("Origin")
	if !strings.Contains(requestOrigin, fmt.Sprintf("http://%s", r.Host)) {
		log.Errorf("Invalid Origin %s", requestOrigin)
		http.Error(w, "Invalid Origin", 403)
		return nil
	}
	conn := s.upgradeToWebsocket(w, r)
	go s.listenToClient(conn)
	return conn
}

func (s *Server) RegisterClientToWriteChannel(client *WebsocketClient, writeChannel *chan Message) bool {
	if _, ok := s.writeMessageChannels[writeChannel]; !ok {
		log.Errorf("Write Channel %s has not been registered", writeChannel)
		return false
	}
	s.writeMessageChannels[writeChannel] = append(s.writeMessageChannels[writeChannel], client)
	return true
}

// Goroutine function
func (s *Server) registerNewChannels() {
	for {
		channel := <-s.RegisterChannelChannel
		s.writeMessageChannels[channel] = make([]*WebsocketClient, 0)
		go s.writeToClients(channel)
	}
}

// Goroutine function
// closeWebsocket is to be invoked when server wants to terminate a websocket
func (s *Server) closeWebsocket() {
	for {
		client := <-s.quitChannel
		log.Debugf("Closing websocket %s", client)
		_ = client.Socket.Close()
		delete(s.clients, client)
	}
}

// Goroutine function
// writeToClients is to be invoked when a channel wishes to write to its clients
func (s *Server) writeToClients(channel *chan Message) {
	for {
		msg := <-*channel
		for _, client := range s.writeMessageChannels[channel] {
			log.Debugf("Writing to a client at %s with message %v", client.Socket.RemoteAddr().String(), msg.Contents)
			err := client.Socket.WriteJSON(msg)
			if err != nil {
				if !s.checkClientClosed(client, err) {
					log.Errorf("Error reading json: %s", err)
				}
			}
		}
	}
}

// Goroutine function
// listenToClient is to be invoked when a database writes a message to the server
// All messages will be broadcast through a single channel
func (s *Server) listenToClient(client *WebsocketClient) {
	for {
		jsonMessage := make(map[string]interface{})
		err := client.Socket.ReadJSON(&jsonMessage)
		if err != nil {
			if s.checkClientClosed(client, err) {
				break
			}
			log.Errorf("Error reading json: %s", err)
			continue
		}
		// Broadcast message
		m := Message{
			Client: client,
			Contents: jsonMessage,
		}
		//log.Debugf("Got message from client: %#v", m)

		s.ClientMessageChannel <- m

		if checkExitMessage(m) {
			s.quitChannel <- client
			break
		}
	}
}

// upgradeToWebsocket upgrades a http call into a websocket
func (s *Server) upgradeToWebsocket(w http.ResponseWriter, r *http.Request) *WebsocketClient {
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		log.Fatal(err)
	}
	log.Debugf("Registering a client at %s", socket.RemoteAddr().String())
	client := WebsocketClient{
		Socket: socket,
	}
	s.clients[&client] = true

	client.Socket.SetCloseHandler(func(code int, text string) error {
		log.Debug("WebsocketClient requesting close")
		s.quitChannel <- &client
		return nil
	})
	return &client
}

// checkClientClosed checks if the database was closed abruptly (no close message was sent to server)
func (s *Server) checkClientClosed(client *WebsocketClient, err error) bool {
	if ce, ok := err.(*websocket.CloseError); ok {
		log.Debugf("WebsocketClient at %s closed with: %s", client.Socket.RemoteAddr().String(), strconv.Itoa(ce.Code))
		s.quitChannel <- client
		return true
	}
	return false
}

// TODO actual exit message
func checkExitMessage(message Message) bool {
	return message.Id == -1
}
