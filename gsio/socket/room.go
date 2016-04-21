package socket

import (
	"com.grid/chsen/gsio/utils"
	"sync"
)

type Room struct {
	uuid    string
	// server reference
	server  *Server
	// name of the room
	Name    string
	// all the clients in the room
	clients map[string]*Client
	mtx     *sync.RWMutex
}

func NewRoom(server *Server, name string) *Room {
	return &Room{
		uuid: 		utils.GenerateUID(),
		server: 	server,
		Name: 		name,
		clients: 	make(map[string]*Client),
		mtx: 		new(sync.RWMutex),
	}
}

func (room *Room) addClient(client *Client) {
	room.mtx.Lock()
	room.clients[client.uuid] = client
	room.mtx.Unlock()
}

func (room *Room) removeClient(client *Client) {
	room.mtx.Lock()
	delete(room.clients, client.uuid)
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

// Disconnects clients from the room
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