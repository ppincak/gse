package socket

import (
	"com.grid/chsen/chsen/utils"
	"sync"
)

type Room struct {
	server  *Server
	uuid    string
	// name of the room
	Name    string
	// all the clients int the room
	clients map[string]*Client
	// mutex
	mtx     *sync.RWMutex
}

func NewRoom(server *Server, name string) *Room {
	return &Room{
		server: 	server,
		uuid: 		utils.GenerateUID(),
		Name: 		name,
		clients: 	make(map[string]*Client),
		mtx: 		new(sync.RWMutex),
	}
}

func (room *Room) addClient(client *Client) {
	room.mtx.Lock()
	room.clients[client.uid] = client
	room.mtx.Unlock()
}

func (room *Room) removeClient(client *Client) {
	room.mtx.Lock()
	delete(room.clients, client.uid)
	room.mtx.Unlock()
}

func (room *Room) GetClients() []*Client {
	room.mtx.Lock()
	clients := make([]*Client, len(room.clients))
	i := 0
	for _, client := range room.clients {
		clients[i] = client
	}
	room.mtx.Unlock()
	return clients
}

func (room *Room) Disconnect() {
	for _, client := range room.clients {
		client.leaveRoom(room)
	}
}

func (room *Room) DestroyRoom() {
	for _, client := range room.clients {
		client.leaveRoom(room)
	}

	room.mtx.Lock()
	room.clients = make(map[string] *Client);
	room.mtx.Unlock()
}

func (room *Room) SendEvent(event string, data interface{}) {
	for _, client := range room.clients {
		client.SendEvent(event, data)
	}
}