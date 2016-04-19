package socket

import (
	"github.com/gorilla/websocket"
	"com.grid/chsen/chsen/utils"
	"encoding/json"
	"sync"
	"com.grid/chsen/chsen/store"
)

type Client struct {
	// pointer to server
	server *Server
	// uid of the room
	uid    string
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

func NewClient(server  *Server, ws *websocket.Conn) (*Client) {
	return &Client{
		server: server,
		uid: 	utils.GenerateUID(),
		ws: ws,
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
		tmsg, err := client.processMessage(msg)
		if err != nil {
			continue
		}

		evt := &event{
			name: tmsg.Event,
			client: client,
			data: tmsg.Data,
		}
		client.server.evc <-evt
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

func (client *Client) JoinRoom(roomName string) {

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

