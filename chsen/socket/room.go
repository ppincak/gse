package socket

import (
	"com.grid/chsen/chsen/utils"
)

type Room struct {
	server  	*Server
	Uid     	string
	// name of the room
	Name    	string
	clients 	map[string]*Client
	brc     	chan *broadcastMessage
	mngc    	chan *management
	stopc   	chan struct{}
}

func NewRoom(server *Server, name string) *Room {
	return &Room{
		server: 	server,
		Uid: 		utils.GenerateUID(),
		Name: 		name,
		clients: 	make(map[string]*Client),
		brc: 		make(chan *broadcastMessage),
		mngc:	make(chan *management),
		stopc: 		make(chan struct{}),
	}
}

func (room *Room) Run() {
	for {
		select {
			case msg := <-room.mngc:
				client := msg.client
				switch msg.msgType {
					case addClient:
						room.clients[client.uid] = client
					case removeClient:
						delete(room.clients, client.uid)
				}
			case br := <- room.brc:
				room.broadcast(br)
			case <- room.stopc:
				return
		}
	}
}

func (room *Room) Stop() {
	room.stopc <- struct{}{}
}

func (room *Room) addClient(client *Client) {
	room.mngc <- &management{
		msgType: addClient,
		client: client,
	}
}

func (room *Room) removeClient(client *Client) {
	room.mngc <- &management{
		msgType: removeClient,
		client: client,
	}
}

func (room *Room) notifyClients() {
	for _, client := range room.clients {
		client.leaveRoom(room.Uid)
	}
}

func (room *Room) broadcast(br *broadcastMessage) {
	for _, client := range room.clients {
		client.sendMessage(br)
	}
}