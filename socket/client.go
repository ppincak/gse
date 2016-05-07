package socket

import (
	"github.com/gorilla/websocket"
	"com.grid/gse/utils"
	"encoding/json"
	"sync"
	"com.grid/gse/store"
	"github.com/sirupsen/logrus"
	"errors"
)


type Client struct {
	// uid of the room
	uuid   		string
	// namespaces to which the client is connected
	namespaces 	map[string] *Namespace
	// rooms to which the client is connected
	rooms  		map[string]*Room
	// storage space
	store		socket.Store
	// webSocket connection
	ws     		*websocket.Conn
	// writer channel
	wc     		chan []byte
	// write mutex
	mtx    		*sync.RWMutex
}

type SocketClient struct {
	*Client
	namespace *Namespace
}

func NewClient(ws *websocket.Conn, store socket.Store) (*Client) {
	return &Client{
		uuid: 		utils.GenerateUID(),
		namespaces: make(map[string]*Namespace),
		rooms: 		make(map[string] *Room),
		store: 		store,
		ws:			ws,
		wc: 		make(chan []byte),
		mtx: 		new(sync.RWMutex),
	};
}

func (client *Client) wrap(Namespace *Namespace) *SocketClient {
	return &SocketClient{
		Client:  	client,
		namespace: 	Namespace,
	}
}

func (client *Client) onPacket(bytes []byte) {
	packet, err := transport.Decode(bytes)
	if err != nil {
		logrus.Error(err)
		return
	}

	switch packet.PacketType {
		case transport.Connect:
			err = client.onConnect(packet)
		case transport.Disconnect:
			err = client.onDisconnect(packet)
		case transport.Event:
			err = client.onEvent(packet)
		case transport.Ack:
			client.onAck(packet)
	}

	if err != nil {
		logrus.Error(err)
	}
}

func (client *Client) on(packet *transport.Packet) (*Namespace, error) {
	if packet.Endpoint == "" {
		return nil, errors.New("Packet missing namespace")
	}
	namespace, ok := client.namespaces[packet.Endpoint]
	if !ok {
		return nil, errors.New("Namespace doesn't exist")
	}
	return namespace, nil
}

func (client *Client) onConnect(packet *transport.Packet) error {
	namespace, err := client.on(packet)
	if err == nil {
		namespace.addClient(client)
	}
	return err
}

func (client *Client) onDisconnect(packet *transport.Packet) error {
	namespace, err := client.on(packet)
	if err == nil {
		namespace.removeClient(client)
	}
	return err
}

func (client *Client) onEvent(packet *transport.Packet) error {
	namespace, err := client.on(packet)
	if err == nil {
		namespace.evc <- &listenerEvent{
			Name:		 	packet.Name,
			Data: 			packet.Args,
			ListenerType: 	eventListener,
		}
	}
	return err
}

func (client *Client) onAck(packet *transport.Packet) {

}
// Pump for reading
func (client *Client) readPump() {
	for {
		_, msg, err := client.ws.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				logrus.Error("Disconnect during read:")
				client.disconnectError(err)
				return
			}
			logrus.Error(err)
		}
		client.onPacket(msg)
	}
}

// Pump for writing
func (client *Client) writePump() {
	for {
		select {
			case msg := <-client.wc:
				if err := client.ws.WriteMessage(websocket.TextMessage, msg); err != nil {
					if websocket.IsUnexpectedCloseError(err) {
						logrus.Error("Disconnect during write")
						client.disconnectError(err)
						return
					}
					logrus.Error(err)
				}
		}
	}
}

func (client *Client) destroy() {
	// leave all rooms
	for _, room := range client.rooms {
		room.removeClient(client)
	}
	// remove from namespaces
	for _, namespace := range client.namespaces {
		namespace.removeClient(client)
	}

	client.namespaces = make(map[string]*Namespace);
	client.rooms = make(map[string]*Room)
}

func (client *Client) Disconnect() {
	client.ws.Close()
	client.destroy()
}

func (client *Client) disconnectError(err error) {
	logrus.Error(err)
	logrus.Infof("Client connection closed, sessionid: %s", client.uuid)
}

func (client *Client) GetSessionId() string {
	return client.uuid
}

func (client *Client) addNamespace(namespace *Namespace) {
	client.mtx.Lock()
	client.namespaces[namespace.name] = namespace
	client.mtx.Unlock()
}

func (n *SocketClient) JoinRoom(roomName string) error {
	room, err := n.namespace.GetRoom(roomName)
    if err != nil {
		return err
	}

	room.addClient(n.Client)
	n.Client.mtx.Lock()
	n.Client.rooms[room.uuid] = room
	n.Client.mtx.Unlock()
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

func (client *Client) disconnectFromNamespaces() {
	for _, namespace := range client.namespaces {
		namespace.removeClient(client)
	}
	client.mtx.Lock()
	client.namespaces = make(map[string]*Namespace);
	client.mtx.Unlock()
}

// send message to client connection
func (client *Client) sendMessage(packet *transport.Packet) {
	raw, err := json.Marshal(packet)
	if err != nil {
		logrus.Debug(Errors[FailedToParsePacket], err)
	}
	client.wc <- raw
}

func (client *Client) sendPacketEvent(event string) {
	packet := &transport.Packet{
		PacketType: transport.Connect,
		Endpoint: "/",
	}
	raw, err := json.Marshal(packet)
	if err != nil {
		logrus.Debug(Errors[FailedToParsePacket], err)
	}
	client.wc <- raw
}

func (client *Client) SendEvent(event string, data interface{}) {
	packet := &transport.Packet{
		Name: event,
		Data: data,
		PacketType: transport.Connect,
	}
	raw, err := json.Marshal(packet)
	if err != nil {
		logrus.Debug(Errors[FailedToParsePacket], err)
	}
	client.wc <- raw
}

// send message to client connection
func (client *Client) SendRaw(data []byte) {
	client.wc <- data
}