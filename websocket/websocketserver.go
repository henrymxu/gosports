package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ngaut/log"
	"net/http"
	"strings"
)

type Server struct {
	ClientMessageChannel   chan Message         // Used for messages received from clients
	RegisterChannelChannel chan RegisterChannel // Used to register new available channels
	quitChannel            chan *Client
	clients                map[*Client]bool //TODO: remove?
	writeMessageChannels   map[*chan Message][]*Client //TODO: threadsafe
	upgrader               websocket.Upgrader
}

type RegisterChannel struct {
	Channel *chan Message
	Action  bool
}

type Client struct {
	Socket  *websocket.Conn // Websocket connection
	Channel *chan Message   // TODO: not used (Channel the websocket is registered to (Only 1 Channel is supported))
}

// Define our message object
type Message struct {
	Client   *Client // Where the message came from
	Type     string
	Id       int
	Contents map[string]interface{}
}

func CreateWebsocketServer() *Server {
	server := Server{
		ClientMessageChannel:   make(chan Message),
		RegisterChannelChannel: make(chan RegisterChannel),
		quitChannel:            make(chan *Client),
		clients:                make(map[*Client]bool),
		writeMessageChannels:   make(map[*chan Message][]*Client),
		upgrader:               websocket.Upgrader{},
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

func (s *Server) BaseWebsocketHandler(w http.ResponseWriter, r *http.Request) *Client {
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

func (s *Server) WriteToClient(client *Client, message Message) {
	log.Debugf("Writing to a client at %s with message type: %s", client.Socket.RemoteAddr().String(), message.Type)
	err := client.Socket.WriteJSON(message)
	if err != nil {
		if !s.checkClientClosed(client, err) {
			log.Errorf("Error reading json: %s", err)
		}
	}
}

func (s *Server) RegisterClientToWriteChannel(client *Client, writeChannel *chan Message) bool {
	if _, ok := s.writeMessageChannels[writeChannel]; !ok {
		log.Errorf("Write channel %v has not been registered", &writeChannel)
		return false
	}
	s.writeMessageChannels[writeChannel] = append(s.writeMessageChannels[writeChannel], client)
	return true
}

// Goroutine function
func (s *Server) registerNewChannels() {
	for {
		registerChannel := <-s.RegisterChannelChannel
		channel := registerChannel.Channel
		if registerChannel.Action {
			log.Debugf("Registered new channel %v", channel)
			s.writeMessageChannels[channel] = make([]*Client, 0)
			go s.writeToClients(channel)
		} else {
			log.Debugf("Deleting channel %v", channel)
			close(*channel)
			delete(s.writeMessageChannels, channel)
		}
	}
}

// Goroutine function
// closeWebsocket is to be invoked when server wants to terminate a websocket
func (s *Server) closeWebsocket() {
	for {
		client := <-s.quitChannel
		log.Debugf("Closing websocket %s", client.Socket.RemoteAddr())
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
			s.WriteToClient(client, msg)
		}
	}
}

// Goroutine function
// listenToClient is to be invoked when a database writes a message to the server
// All messages will be broadcast through a single channel
func (s *Server) listenToClient(client *Client) {
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
			Client:   client,
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
func (s *Server) upgradeToWebsocket(w http.ResponseWriter, r *http.Request) *Client {
	socket, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		log.Fatal(err)
	}
	log.Debugf("Registering a client at %s", socket.RemoteAddr().String())
	client := Client{
		Socket: socket,
	}
	s.clients[&client] = true

	client.Socket.SetCloseHandler(func(code int, text string) error {
		log.Debug("Client requesting close")
		s.quitChannel <- &client
		return nil
	})
	return &client
}

// checkClientClosed checks if the database was closed abruptly (no close message was sent to server)
func (s *Server) checkClientClosed(client *Client, err error) bool {
	if ce, ok := err.(*websocket.CloseError); ok {
		log.Debugf("Client at %s closed with: %s", client.Socket.RemoteAddr().String(), ce.Text)
		s.quitChannel <- client
		return true
	}
	return false
}

// TODO actual exit message
func checkExitMessage(message Message) bool {
	return message.Id == -1
}
