package socket

import (
	"com.grid/chsen/gsio/utils"
	"sync"
)

type Room struct {
	uuid   		string
	// name of the room
	name	   	string
	// namespace reference
	namespace	*Namespace
	// all the clients in the room
	clients 	map[string]*Client
	// room lock
	mtx     	*sync.RWMutex
}

func NewRoom(namespace *Namespace, name string) *Room {
	return &Room{
		uuid: 		utils.GenerateUID(),
		name: 		name,
		namespace: 	namespace,
		clients: 	make(map[string]*Client),
		mtx: 		new(sync.RWMutex),
	}
}

func (room *Room) GetSessionId() string {
	return room.uuid
}

func (room *Room) GetName() string {
	return room.name
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