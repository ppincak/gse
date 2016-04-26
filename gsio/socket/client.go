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
	uuid   		string
	// pointer to server
	namespace 	*Namespace
	// websocket connection
	ws     		*websocket.Conn
	// storage space
	store  		socket.Store
	// rooms to which the client is connected
	rooms  		map[string]*Room
	// writer channel
	wc     		chan []byte
	// write mutex
	mtx    		*sync.RWMutex
}

func NewClient(namespace  *Namespace, ws *websocket.Conn, store socket.Store) (*Client) {
	return &Client{
		uuid: 		utils.GenerateUID(),
		namespace: 	namespace,
		ws:			ws,
		store: 		store,
		rooms: 		make(map[string] *Room),
		wc: 		make(chan []byte),
		mtx: 		new(sync.RWMutex),
	};
}

func(client *Client) processMessage(rawMsg []byte) (*TransportMessage, error) {
	var tmsg TransportMessage
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
				client.disconnectError(err)
				return
			}
		}
		tmsg, err := client.processMessage(msg)
		if err != nil {
			continue
		}

		evt := &Event{
			EventType: eventListener,
			Name: tmsg.Event,
			Client: client,
			Data: tmsg.Data,
		}
		client.namespace.evc <-evt
	}
}

// Pump for writing
func (client *Client) writePump() {
	for {
		select {
			case msg := <-client.wc:
				if err := client.ws.WriteMessage(websocket.TextMessage, msg); err != nil {
					client.disconnectError(err)
					return
				}
		}
	}
}

func (client *Client) Disconnect() {
	client.leaveAllRooms()
	client.namespace.removeClient(client)
	client.ws.Close()
}

func (client *Client) disconnectError(err error) {
	client.Disconnect()
	logrus.Error(err)
	client.namespace.statIncr(FailedConnections)
}

func (client *Client) GetSessionId() string {
	return client.uuid
}

func (client *Client) JoinRoom(roomName string) error {
	room, err := client.namespace.GetRoom(roomName)
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
		if room.name == roomName {
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

// send message to client connection
func (client *Client) sendMessage(tmsg *TransportMessage) {
	raw, err := json.Marshal(tmsg)
	if err != nil {
		logrus.Debug(Errors[FailedToParseMessage], err)
	}
	client.wc <- raw
}

func (client *Client) SendEvent(event string, data interface{}) {
	tmsg := &TransportMessage{
		Event: event,
		Data: data,
	}
	raw, err := json.Marshal(tmsg)
	if err != nil {
		logrus.Debug(Errors[FailedToParseMessage], err)
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