package socket

import (
	"github.com/gorilla/websocket"
	"com.grid/chsen/gsio/utils"
	"encoding/json"
	"sync"
	"com.grid/chsen/gsio/store"
	"github.com/sirupsen/logrus"

)

type Client struct {
	// uid of the room
	uuid   string
	// pointer to server
	server *Server
	// websocket connection
	ws     *websocket.Conn
	// storage space
	store  socket.Store
	// rooms to which the client is connected
	rooms  map[string]*Room
	// writer channel
	wc     chan []byte
	// write mutex
	mtx    *sync.RWMutex
}

func NewClient(server  *Server, ws *websocket.Conn, store socket.Store) (*Client) {
	return &Client{
		uuid: 	utils.GenerateUID(),
		server: server,
		ws: ws,
		store: store,
		rooms: 	make(map[string] *Room),
		wc: make(chan []byte),
		mtx: new(sync.RWMutex),
	};
}

func(client *Client) processMessage(rawMsg []byte) (*transportmessage, error) {
	var tmsg transportmessage
	err := json.Unmarshal(rawMsg, &tmsg)
	if err != nil {
		return nil, err
	}
	return &tmsg, nil
}

// Pump for reading
func (client *Client) readPump() {
	for {
		_, msg, err := client.ws.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) ||
			   websocket.IsUnexpectedCloseError(err) {
				logrus.Errorln("Client disconnected", err)
				client.Disconnect()
				return
			}
		}
		tmsg, err := client.processMessage(msg)
		if err != nil {
			continue
		}

		evt := &Event{
			EventType: Custom,
			Name: tmsg.Event,
			Client: client,
			Data: tmsg.Data,
		}
		client.server.evc <-evt
	}
}

// Pump for writing
func (client *Client) writePump() {
	for {
		select {
			case msg := <-client.wc:
				if err := client.ws.WriteMessage(websocket.TextMessage, msg); err != nil {
					client.Disconnect()
					return
				}
		}
	}
}

func (client *Client) Disconnect() {
	client.leaveAllRooms()
	client.server.removeClient(client)
	client.ws.Close()
}

func (client *Client) GetSessionId() string {
	return client.uuid
}

func (client *Client) JoinRoom(roomName string) error {
	room, err := client.server.getRoom(roomName)
    if err != nil {
		return err
	}
	room.addClient(client)
	client.mtx.Lock()
	client.rooms[room.uuid] = room
	client.mtx.Unlock()
	return nil
}

func (client *Client) joinRoom(room *Room) {
	client.mtx.Lock()
	client.rooms[room.uuid] = room
	client.mtx.Unlock()
	room.addClient(client)
}

func (client *Client) LeaveRoom(roomName string) {
	for _, room := range client.rooms {
		if room.Name == roomName {
			client.leaveRoom(room)
			return
		}
	}
}

func (client *Client) leaveRoom(room *Room) {
	client.mtx.Lock()
	delete(client.rooms, room.uuid);
	client.mtx.Unlock()
	room.removeClient(client)
}

func (client *Client) leaveAllRooms() {
	for _, room := range client.rooms {
		room.removeClient(client)
	}

	client.mtx.Lock()
	client.rooms = make(map[string]*Room)
	client.mtx.Unlock()
}

func (client *Client) connectionError() {
	logrus.Error()
}

// send message to client connection
func (client *Client) sendMessage(tmsg *transportmessage) {
	raw, err := json.Marshal(tmsg)
	if err != nil {
		// log error
	}
	client.wc <- raw
}

func (client *Client) SendEvent(event string, data interface{}) {
	tmsg := &transportmessage{
		Event: event,
		Data: data,
	}
	raw, err := json.Marshal(tmsg)
	if err != nil {
		// log error
	}
	client.wc <- raw
}

// send message to client connection
func (client *Client) SendRaw(data []byte) {
	client.wc <- data
}

func (client *Client) Set(key string, value interface{}) {
	client.store.Set(key, value)
}

func (client *Client) Get(key string) interface{} {
	return client.store.Get(key)
}

func (client *Client) Delete(key string) {
	client.store.Delete(key)
}

func (client *Client) Has(key string) bool {
	return client.store.Has(key)
}