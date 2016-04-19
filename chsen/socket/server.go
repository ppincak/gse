package socket

import (
	"net/http"
	"github.com/gorilla/websocket"
	"com.grid/chsen/chsen/conf"
	"github.com/sirupsen/logrus"
	"com.grid/chsen/chsen/analytics"
	"sync"
	"errors"
)

const(
	ReadBufferSize = 1024
	WriteBufferSize = 1024
)

type Server struct {
	upgrader  *websocket.Upgrader
	// map of all connected rooms
	rooms     map[string]*Room
	// map of all connected clients
	clients   map[string]*Client
	// map of all server listeners
	listeners map[string] []EventCallback
	// broadcast channel
	brdc      chan *transportmessage
	// events channel
	evc       chan *event
	// listeners registration channel
	liregc    chan listenerMessage
	// registration channel
	regc      chan *Client
	// un registration channel
	unregc    chan *Client
	// channel for providing status
	statc     chan chan Status
	// channel for server stopping
	stopc     chan struct{}
	// mutex
	mtx		  *sync.RWMutex


	// server configuration
	conf      *conf.Conf
	// server stats
	stats     *analytics.Stats
}

type Status struct {
	ServerName  string		`json:"serverName"`
	Clients 	int			`json:"clients"`
	Rooms   	int			`json:"rooms"`
}

func NewServer(config *conf.Conf) *Server {
	if config == nil {
		config = conf.DefaultConf()
	}

	upgrader := &websocket.Upgrader{
		ReadBufferSize: 	ReadBufferSize,
		WriteBufferSize: 	WriteBufferSize,
	}

	return &Server {
		upgrader: 	upgrader,
		rooms: 		make(map[string]*Room),
		clients: 	make(map[string]*Client),
		listeners: 	make(map[string] []EventCallback),
		brdc: 		make(chan *transportmessage),
		evc: 		make(chan *event),
		liregc: 	make(chan listenerMessage),
		statc: 		make(chan chan Status),
		regc:  		make(chan *Client),
		unregc: 	make(chan *Client),
		stopc: 		make(chan struct{}),
		mtx: 		new(sync.RWMutex),
		stats: 		analytics.NewStats(),
		conf: 		config,
	}
}

func (server *Server) Run() {
	for {
		select {
			case client := <- server.regc:
				server.clients[client.uuid] = client
			case client := <- server.unregc:
				delete(server.clients, client.uuid)
			case msg := <- server.liregc:
				if listeners, ok := server.listeners[msg.event]; ok {
					server.listeners[msg.event] = append(listeners, msg.callback)
				} else {
					server.listeners[msg.event] = []EventCallback {msg.callback}
				}
			case evt := <- server.evc:
				if listeners, ok := server.listeners[evt.name]; ok{
					for _, listener := range listeners {
						listener(evt.client, evt.data)
					}
				}
				return
			case msg := <- server.brdc:
				for _, client := range server.clients {
					client.sendMessage(msg)
				}
			case c := <- server.statc:
				c <- server.getStatus()
			case <- server.stopc:
				return
		}
	}
}

func (server *Server) Stop() {
	server.stopc <- struct{}{}
}

func (server *Server) Listen(event string, callback EventCallback) {
	server.liregc <- listenerMessage{
		event: event,
		callback: callback,
	}
}

func (server *Server) getStatus() (Status) {
	return Status{
		Clients: len(server.clients),
		Rooms: len(server.rooms),
	}
}

func (server *Server) getStats() (analytics.Stats) {
	return *server.stats
}

func (server *Server) createRoom(roomName string) string {
	server.mtx.Lock()
	room := NewRoom(server, roomName)
	server.rooms[room.uuid] = room
	server.mtx.Unlock()
	return room.uuid
}

func (server *Server) getRoom(roomName string) (*Room, error) {
	server.mtx.RLock()
	for _, room := range server.rooms {
		if room.Name == roomName {
			server.mtx.RUnlock()
			return room, nil
		}
	}
	server.mtx.RUnlock()
	return nil, errors.New("Room not found")
}

func (server *Server) deleteRoom(roomName string) {
	server.mtx.Lock()
	for _, room := range server.rooms {
		if room.Name == roomName {
			room.DestroyRoom()
			delete(server.rooms, roomName)
			return
		}
	}
	server.mtx.Unlock()
}

func (server *Server) joinRoom(roomName string, client *Client) {
	server.mtx.RLock()
	for _, room := range server.rooms {
		if room.Name == roomName {
			room.addClient(client)
			client.joinRoom(room)
			server.mtx.RUnlock()
			return
		}
	}
	server.mtx.RUnlock()
}

func (server *Server) addClient(client *Client) {
	server.regc <- client
	server.stats.Inc(analytics.OpenedClients)
}

func (server *Server) removeClient(client *Client) {
	server.unregc <- client
	server.stats.Inc(analytics.ClosedClients)
}

func (server *Server) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error(err)
		return
	}
	server.addClient(NewClient(server, ws))
}