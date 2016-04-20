package socket

import (
	"net/http"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"sync"
	"errors"
)

const(
	initialLSize = 5

	ReadBufferSize = 1024
	WriteBufferSize = 1024
)

type Server struct {
	upgrader   *websocket.Upgrader
	// map of all connected rooms
	rooms      map[string]*Room
	// map of all connected clients
	clients    map[string]*Client
	// event listeners
	listeners  map[string] []EventCallback
	//  connection callbacks
	cListeners []ConnectCallback
	// disconecct listeners
	dListeners []DisconnectCallback
	// transport channel
	trpc       chan *transportmessage
	// events channel
	evc        chan *event
	// event listeners registration channel
	liregc     chan listenerMessage
	// connection callbacks registration channel
	clregc     chan ConnectCallback
	//
	dlregc     chan DisconnectCallback
	// channel for server stopping
	stopc      chan struct{}
	// data mutex
	mtx        *sync.RWMutex
	// server configuration
	conf       *Conf
	// server stats
	stats      *Stats
}

func NewServer(config *Conf) *Server {
	if config == nil {
		config = DefaultConf()
	}

	upgrader := &websocket.Upgrader{
		ReadBufferSize: 	ReadBufferSize,
		WriteBufferSize: 	WriteBufferSize,
	}

	return &Server {
		upgrader: 		upgrader,
		rooms: 			make(map[string]*Room),
		clients: 		make(map[string]*Client),
		listeners: 		make(map[string] []EventCallback),
		cListeners:     make([]ConnectCallback, initialLSize),
		dListeners:     make([]DisconnectCallback, initialLSize),
		trpc: 			make(chan *transportmessage),
		evc: 			make(chan *event),
		liregc: 		make(chan listenerMessage),
		clregc:      	make(chan ConnectCallback),
		dlregc:      	make(chan DisconnectCallback),
		stopc: 			make(chan struct{}),
		mtx: 			new(sync.RWMutex),
		stats: 			NewStats(),
		conf: 			config,
	}
}

func (server *Server) Run() {
	for {
		select {
			case msg := <- server.liregc:
				if listeners, ok := server.listeners[msg.event]; ok {
					server.listeners[msg.event] = append(listeners, msg.callback)
				} else {
					server.listeners[msg.event] = []EventCallback {msg.callback}
				}
			case callback := <- server.clregc:
				server.cListeners = append(server.cListeners, callback)
			case callback := <- server.dlregc:
				server.dListeners = append(server.dListeners, callback)
			case evt := <- server.evc:
				if listeners, ok := server.listeners[evt.name]; ok{
					for _, listener := range listeners {
						listener(evt.client, evt.data)
					}
				}
				return
			case msg := <- server.trpc:
				for _, client := range server.clients {
					client.sendMessage(msg)
				}
			case <- server.stopc:
				return
		}
	}
}

func (server *Server) Stop() {
	server.stopc <- struct{}{}
}

func (server *Server) AddConnectListener(callback ConnectCallback) {
	server.clregc <- callback
}

func (server *Server) AddDisconnectListener(callback DisconnectCallback) {
	server.dlregc <- callback
}

func (server *Server) Listen(event string, callback EventCallback) {
	server.liregc <- listenerMessage{
		event: event,
		callback: callback,
	}
}

func (server *Server) SendEvent(event string, data interface{}) {
	server.trpc <- &transportmessage{
		Event: event,
		Data: data,
	}
}

func (server *Server) getStatus() (Status) {
	server.mtx.RLock()
	defer server.mtx.RUnlock()
	numRooms := len(server.rooms)
	roomStatus := make([]RoomStatus, numRooms)

	i := 0
	for _, room := range server.rooms {
		roomStatus[i] = RoomStatus{
			RoomName: room.Name,
			NumberOfClients: len(room.clients),
		}
		i++
	}
	server.mtx.RUnlock()

	return Status{
		ServerName: server.conf.ServerName,
		NumberOfClients: len(server.clients),
		NumberOfRooms: numRooms,
		RoomsStatus: roomStatus,
	}
}

func (server *Server) getStats() (Stats) {
	return *server.stats
}

func (server *Server) addRoom(roomName string) string {
	server.mtx.Lock()
	room := NewRoom(server, roomName)
	server.rooms[room.uuid] = room
	server.mtx.Unlock()
	server.stats.Inc(OpenedRooms)
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

func (server *Server) removeRoom(roomName string) {
	server.mtx.Lock()
	defer server.mtx.Unlock()
	for _, room := range server.rooms {
		if room.Name == roomName {
			room.DestroyRoom()
			delete(server.rooms, roomName)
			server.mtx.Unlock()
			server.stats.Inc(ClosedRooms)
			return
		}
	}
	server.mtx.Unlock()
}

func (server *Server) joinRoom(roomName string, client *Client) {
	server.mtx.RLock()
	defer server.mtx.RUnlock()
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
	server.mtx.Lock()
	server.clients[client.uuid] = client
	server.mtx.Unlock()
    server.stats.Inc(OpenedClients)
}

func (server *Server) removeClient(client *Client) {
	server.mtx.Lock()
	delete(server.clients, client.uuid)
	server.mtx.Unlock()
	server.stats.Inc(ClosedClients)
}

func (server *Server) serveWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error(err)
		return
	}
	server.addClient(NewClient(server, ws))
}