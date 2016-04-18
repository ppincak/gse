package socket

import (
	"github.com/gorilla/websocket"
	"com.grid/chsen/chsen/utils"
	"errors"
	"encoding/json"
	"com.grid/chsen/chsen/stats"

	"sync"
)

type Processor interface {
	Process(in <-chan []byte, out chan<- []byte)
}

type Client struct {
	// pointer to server
	server *Server
	// uid of the room
	uid    string
	// websocket connection
	ws     *websocket.Conn
	// storage space
	store  map[string]interface{}
	// rooms to which the client is connected
	rooms  map[string]*Room
	// management channel
	mngc   chan *management
	// reader channel
	rc     chan []byte
	// writer channel
	wc     chan []byte

	mtx    *sync.RWMutex

	proc   Processor
}

func NewClient(server  *Server, ws *websocket.Conn) (*Client) {
	return &Client{
		server: server,
		uid: 	utils.GenerateUID(),
		ws: ws,
		rooms: 	make(map[string] *Room),
		mngc: make(chan *management),
		rc: make(chan []byte),
		wc: make(chan []byte),
		mtx: new(sync.RWMutex),
	};
}

func (client *Client) Run() {
	go client.readPump()
	go client.writePump()

	for {
		select {
			case msg := <- client.rc:
				msg = msg
			}
	}
}

// pump for reading
func (client *Client) readPump() {
	for {
		_, msg, err := client.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				client.destroyClient()
				return
			}
		}
		client.rc <-msg
	}
}

// pump for writing
func (client *Client) writePump() {
	for {
		select {
			case msg := <-client.wc:
				if err := client.ws.WriteMessage(websocket.TextMessage, msg); err != nil {
					client.destroyClient()
					return
				}
		}
	}
}

func (client *Client) destroyClient() {
	client.leaveAllRooms()
	client.server.removeClient(client)
	client.ws.Close()
}

func (client *Client) joinRoom(room *Room) {
	client.mtx.Lock()
	client.rooms[room.Uid] = room
	client.mtx.Unlock()
}

func (client *Client) leaveRoom(roomId string) {
	client.mtx.Lock()
	delete(client.rooms, roomId);
	client.mtx.Unlock()
}

func (client *Client) leaveAllRooms() {
	for _, room := range client.rooms {
		room.removeClient(client)
	}

	client.mtx.Lock()
	client.rooms = make(map[string]*Room)
	client.mtx.Unlock()
}

func (client *Client) set(key string, value interface{}) {
	client.mtx.Lock()
	client.store[key] = value
	client.mtx.Unlock()
}

func (client *Client) get(key string) interface{} {
	client.mtx.RLock()
	val := client.store[key]
	client.mtx.RUnlock()
	return val
}

func (client *Client) has(key string) bool {
	client.mtx.RLock()
	_, ok := client.store[key]
	client.mtx.RUnlock()
	return ok
}

func (client *Client) del(key string) {
	client.mtx.Lock()
	delete(client.store, key)
	client.mtx.Unlock()
}

// send message to client connection
func (client *Client) Send(data []byte) {
	client.wc <- data
}

// send message to client connection
func (client *Client) sendMessage(msg *broadcastMessage) {
	srl, err := json.Marshal(msg)
	if err != nil {
		client.server.stats.Inc(stats.FailedMessages)
	}
	client.wc <- srl
}