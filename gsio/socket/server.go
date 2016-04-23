package socket

import (
	"net/http"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"sync"
	"errors"
	"fmt"

	"com.grid/chsen/gsio/store"
)

type listeners struct {
	listeners  map[string] []EventCallback
	//  connection callbacks
	cListeners []ConnectCallback
	// disconecct listeners
	dListeners []DisconnectCallback
}

type Server struct {
	upgrader   		*websocket.Upgrader
	// map of all connected rooms
	rooms      		map[string]*Room
	// map of all connected clients
	clients    		map[string]*Client
	// event listeners
	listeners    	map[string] []EventCallback
	//  connection callbacks
	cListeners   	[]ConnectCallback
	// disconecct listeners
	dListeners   	[]DisconnectCallback
	// transport channel
	trpc         	chan *transportmessage
	// events channel
	evc          	chan *Event
	// event listeners registration channel
	liregc       	chan registerListener
	// channel for server stopping
	stopc        	chan struct{}
	// data mutex
	mtx          	*sync.RWMutex
	// server configuration
	conf         	*ServerConf
	//  store factory
	storeFactory 	socket.LocalStoreFactory
	// server stats
	stats        	*Stats
}

func NewServer(storeFactory socket.LocalStoreFactory, config *ServerConf) *Server {
	if storeFactory == nil {
		storeFactory = socket.NewLocalStore
	}
	if config == nil {
		config = DefaultConf()
	}

	upgrader := &websocket.Upgrader{
		ReadBufferSize: 	config.ReadBufferSize,
		WriteBufferSize: 	config.WriteBufferSize,
	}

	return &Server {
		upgrader: 		upgrader,
		rooms: 			make(map[string] *Room),
		clients: 		make(map[string] *Client),
		listeners: 		make(map[string] []EventCallback),
		cListeners:     make([]ConnectCallback, 0),
		dListeners:     make([]DisconnectCallback, 0),
		trpc: 			make(chan *transportmessage),
		evc: 			make(chan *Event),
		liregc: 		make(chan registerListener),
		stopc: 			make(chan struct{}),
		mtx: 			new(sync.RWMutex),
		stats: 			NewStats(),
		storeFactory: 		storeFactory,
		conf: 			config,
	}
}

func (server *Server) Run() {
	logrus.Info("Starting socket server worker")

	for {
		select {
			case msg := <- server.liregc:
				logrus.Infof("Registering event: %+v", msg)
				switch msg.listenerType {
					case Connect:
						server.cListeners = append(server.cListeners, msg.ConnectCallback)
					case Disconnect:
						server.dListeners = append(server.dListeners, msg.DisconnectCallback)
					case Custom:
						if listeners, ok := server.listeners[msg.event]; ok {
							server.listeners[msg.event] = append(listeners, msg.EventCallback)
						} else {
							server.listeners[msg.event] = []EventCallback {msg.EventCallback}
						}
				}
			case evt := <- server.evc:
				switch evt.EventType {
					case Connect:
						fmt.Println( server.cListeners)
						for _, callback := range server.cListeners{
							go callback(evt.Client)
						}
					case Disconnect:
						for _, callback := range server.dListeners{
							go callback(evt.Client)
						}
					case Custom:
						if listeners, ok := server.listeners[evt.Name]; ok{
							for _, listener := range listeners {
								go listener(evt.Client, evt.Data)
							}
						}
				}
			case msg := <- server.trpc:
				logrus.Info("Sending message to all clients")
				for _, client := range server.clients {
					client.sendMessage(msg)
				}
			case <- server.stopc:
				logrus.Info("Socket server worker stopped")
				return
		}
	}
}

func (server *Server) Stop() {
	logrus.Info("Stopping socket server worker")
	server.stopc <- struct{}{}
}

func (server *Server) AddConnectListener(callback ConnectCallback) {
	server.liregc <- registerListener{
		listenerType: Connect,
		ConnectCallback: callback,
	}
}

func (server *Server) AddDisconnectListener(callback DisconnectCallback) {
	server.liregc <- registerListener{
		listenerType: Disconnect,
		DisconnectCallback: callback,
	}
}

func (server *Server) Listen(event string, callback EventCallback) {
	server.liregc <- registerListener{
		listenerType: Custom,
		event: event,
		EventCallback: callback,
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

func (server *Server) AddRoom(roomName string) string {
	server.mtx.Lock()
	room := NewRoom(server, roomName)
	server.rooms[room.uuid] = room
	server.mtx.Unlock()
	server.stats.Inc(OpenedRooms)
	return room.uuid
}

func (server *Server) GetRoom(roomName string) (*Room, error) {
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

func (server *Server) RemoveRoom(roomName string) {
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
	server.evc <- &Event{
		EventType: Connect,
		Client: client,
	}
}

func (server *Server) removeClient(client *Client) {
	server.mtx.Lock()
	delete(server.clients, client.uuid)
	server.mtx.Unlock()
	server.stats.Inc(ClosedClients)
	server.evc <- &Event{
		EventType: Disconnect,
		Client: client,
	}
}

func (server *Server) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error(err)
		return
	}
	client := NewClient(server, ws, server.storeFactory())
	go client.readPump()
	go client.writePump()
	server.addClient(client)
}