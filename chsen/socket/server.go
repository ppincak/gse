package socket

import (
	"net/http"
	"github.com/gorilla/websocket"
	"com.grid/chsen/chsen/stats"
	"com.grid/chsen/chsen/conf"
	"github.com/sirupsen/logrus"
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
	brdc      chan *broadcastMessage
	// channel for outside communication
	Cmdc      <-chan *Command
	// events channel
	evc       chan *event
	// management channel
	mngc      chan *management
	// channel for providing status
	statc     chan chan Status
	// channel for server stopping
	stopc     chan struct{}

	// server configuration
	conf      *conf.Conf
	// server stats
	stats     *stats.Stats
}

type Status struct {
	Clients 	int		`json:"clients"`
	Rooms   	int		`json:"rooms"`
}

func NewServer(conf *conf.Conf) *Server {
	if conf == nil {

	}

	upgrader := &websocket.Upgrader{
		ReadBufferSize: ReadBufferSize,
		WriteBufferSize: WriteBufferSize,
	}

	return &Server {
		upgrader: upgrader,
		rooms: make(map[string]*Room),
		clients: make(map[string]*Client),
		brdc: make(chan *broadcastMessage),
		Cmdc: make(chan *Command),
		mngc: make(chan *management),
		statc: make(chan chan Status),
		stopc: make(chan struct{}),
		stats: stats.NewStats(),
		conf: conf,
	}
}

func (server *Server) Run() {
	for {
		select {
			case cmd := <- server.Cmdc:
				roomName := cmd.Data.(string)

				switch cmd.CmdType {
					case CreateRoom:
						room := NewRoom(server, roomName)
						go room.Run()
						server.rooms[room.Uid] = room
						/*cmd.OutputC <- &ServerCommand {

						}*/
					case DestroyRoom:
						if room, ok := server.rooms[cmd.Data]; ok {
							room.notifyClients()
							room.Stop()
							delete(server.rooms, roomName)
						}
				}
			case mng := <- server.mngc:
				client := mng.client
				switch mng.msgType {
					case addClient:
						server.clients[client.uid] = client
					case removeClient:
						delete(server.clients, client.uid)
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

}

func (server *Server) getStatus() (Status) {
	return Status{
		Clients: len(server.clients),
		Rooms: len(server.rooms),
	}
}

func (server *Server) getStats() (stats.Stats) {
	return *server.stats
}

func (server *Server) addClient(client *Client) {
	server.mngc <- &management{
		msgType: addClient,
		client: client,
	}
	server.stats.Inc(stats.OpenedClients)
}

func (server *Server) removeClient(client *Client) {
	server.mngc <- &management{
		msgType: removeClient,
		client: client,
	}
	server.stats.Inc(stats.ClosedClients)
}

func (server *Server) serveWebsocket(w http.ResponseWriter, r *http.Request) {
	ws, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error(err)
		return
	}
	server.addClient(NewClient(server, ws))
}