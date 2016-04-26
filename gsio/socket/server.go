package socket

import (
	"net/http"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"errors"
	"com.grid/chsen/gsio/store"
	"com.grid/chsen/gsio/socket/monitor"
)

type Server struct {
	// root namespace
	*Namespace
	// map of all namespaces
	namespaces		map[string]*Namespace
	// gorilla websocket upgrader
	upgrader   		*websocket.Upgrader
	// channel for server stopping
	stopc        	chan struct{}
	// server configuration
	conf         	*ServerConf
	//  store factory
	storeFactory 	socket.LocalStoreFactory
	// server stats
	stats        	*monitor.Stats

	isRunning		bool
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

	server := &Server {
		upgrader: 		upgrader,
		namespaces:     make(map[string] *Namespace),
		stopc: 			make(chan struct{}),
		stats: 			monitor.NewStats(),
		storeFactory: 	storeFactory,
		conf: 			config,
	}
	server.Namespace = rootNamespace(server)
	return server
}

func (server *Server) AddNamespace(namespaceName string) (*Namespace, error) {
	if server.isRunning {
		return nil, makeError(10)
	}

	namespace := newNamespace(namespaceName, server)
	server.namespaces[namespaceName] = namespace
	return namespace, nil
}

func (server *Server) getStatus() (Status) {
	server.mtx.RLock()
	defer server.mtx.RUnlock()
	numRooms := len(server.rooms)
	roomStatus := make([]RoomStatus, numRooms)

	i := 0
	for _, room := range server.rooms {
		roomStatus[i] = RoomStatus{
			RoomName: room.name,
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

func (server *Server) getStats() (monitor.Stats) {
	return *server.stats
}

func (server *Server) addClient(client *Client, namespaceName string) error {
	namespace, ok := server.namespaces[namespaceName]
	if !ok {
		return errors.New("error")
	}
	namespace.addClient(client)
	return nil
}

func (server *Server) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error(err)
		return
	}

	client := NewClient(server, ws, server.storeFactory())

	err = server.addClient(client, "")
	if err != nil {
		logrus.Error(err)
		return
	}

	go client.readPump()
	go client.writePump()
}