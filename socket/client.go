package socket

import (
	"github.com/ppincak/gse/store"
	"github.com/ppincak/gse/utils"
	"github.com/ppincak/gse/socket/transport"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"encoding/json"
	"sync"
	"errors"
)

type Client struct {
	// uid of the room
	uuid   		string
	// server
	server		*Server
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
	// stop channel
	stopc		chan struct{}
	// write mutex
	mtx    		*sync.RWMutex
	// flag indicating if the connection is open
	open   		bool
}

func NewClient(server *Server, ws *websocket.Conn, store socket.Store) (*Client) {
	return &Client{
		uuid: 		utils.GenerateUID(),
		namespaces: make(map[string]*Namespace),
		server:     server,
		rooms: 		make(map[string] *Room),
		store: 		store,
		ws:			ws,
		wc: 		make(chan []byte),
		stopc:      make(chan struct{}),
		mtx: 		new(sync.RWMutex),
		open:		true,
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
			err = client.onAck(packet)
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
	if packet.Endpoint == "" {
		return errors.New("Packet missing namespace")
	}
	_, ok := client.namespaces[packet.Endpoint]
	if ok {
		return errors.New("Already connected to namespace")
	}
	namespace, ok := client.server.namespaces[packet.Endpoint]
	if !ok {
		return errors.New("Namespace doesn't exist")
	}
	namespace.addClient(client)
	client.addNamespace(namespace)
	return nil
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
		namespace.evc <- client.makeEvent(packet)
	}
	return err
}

func (client *Client) makeEvent(packet *transport.Packet) *listenerEvent {
	return &listenerEvent{
		client: 		client,
		mame:		 	packet.Name,
		data: 			packet.Data,
		listenerType: 	eventListener,
	}
}

func (client *Client) onAck(packet *transport.Packet) error {
	namespace, err := client.on(packet)
	if packet.Id == 0 {
		return errors.New("bad packet id generated on the client side")
	}
	if err == nil {
		event := client.makeEvent(packet)
		event.ack = &Ack{
			id: 		packet.Id,
			client: 	client,
			namespace:  namespace,
		}
		namespace.evc <- event
	}
	return nil
}

func (client *Client) readPump() {
	logrus.Infof("Client: %s readpump started", client.uuid)
	defer logrus.Infof("Client: %s readpump stopped", client.uuid)

	for {
		_, msg, err := client.ws.ReadMessage()

		if err != nil {
			client.disconnectError(err)
			client.stopc <- struct{}{}
			return
		}
		client.onPacket(msg)
	}
}

func (client *Client) writePump() {
	logrus.Infof("Client: %s writepump started", client.uuid)
	defer logrus.Infof("Client: %s writepump stopped", client.uuid)

	for {
		select {
			case msg := <-client.wc:
				if err := client.ws.WriteMessage(websocket.TextMessage, msg); err != nil {
					client.disconnectError(err)
					return
				}
			case <- client.stopc:
				return
		}
	}
}

func (client *Client) isOpen() bool {
	client.mtx.RLock()
	defer client.mtx.RUnlock()
	return client.open
}

func (client *Client) close() {
	client.mtx.Lock()
	client.open = false
	client.mtx.Unlock()
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

	client.server.removeClient(client)

	client.namespaces = make(map[string]*Namespace);
	client.rooms = make(map[string]*Room)
}

func (client *Client) Disconnect() {
	client.ws.Close()
	client.close()
	client.destroy()
}

func (client *Client) disconnectError(err error) {
	client.close()
	client.destroy()
	logrus.Error(err)
	logrus.Errorf("Client connection closed, sessionid: %s", client.uuid)
}

func (client *Client) Store() socket.Store {
	return client.store
}

func (client *Client) GetSessionId() string {
	return client.uuid
}

func (client *Client) addNamespace(namespace *Namespace) {
	client.mtx.Lock()
	client.namespaces[namespace.name] = namespace
	client.mtx.Unlock()
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

func (client *Client) GetAllRooms() []*Room {
	rooms := make([]*Room, len(client.rooms))

	i := 0
	for _, room := range client.rooms {
		rooms[i] = room
		i++
	}

	return rooms
}

func (client *Client) disconnectFromNamespaces() {
	for _, namespace := range client.namespaces {
		namespace.removeClient(client)
	}
	client.mtx.Lock()
	client.namespaces = make(map[string]*Namespace);
	client.mtx.Unlock()
}

func (client *Client) notify(pType transport.PacketType, namespaceName string) {
	client.SendPacket(&transport.Packet{
		PacketType: pType,
		Endpoint:	namespaceName,
	})
}

func (client *Client) SendPacket(packet *transport.Packet) {
	if !client.isOpen() {
		return
	}

	raw, err := json.Marshal(packet)
	if err != nil {
		logrus.Debug(Errors[FailedToParsePacket], err)
	}
	client.wc <- raw
}

func (client *Client) sendEvent(event string, data interface{}, namespaceName string) {
	if !client.isOpen() {
		return
	}

	packet := &transport.Packet{
		Name: event,
		Data: data,
		PacketType: transport.Event,
		Endpoint: namespaceName,
	}
	raw, err := json.Marshal(packet)
	if err != nil {
		logrus.Debug(Errors[FailedToParsePacket], err)
	}
	client.wc <- raw
}

func (client *Client) SendRaw(data []byte) {
	if client.isOpen() {
		client.wc <- data
	}
}

type SocketClient struct {
	*Client
	ack 	  *Ack
	namespace *Namespace
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

func (client *SocketClient) Disconnect() {
	client.namespace.removeClient(client.Client)
	delete(client.namespaces, client.namespace.name)
}

func (client *SocketClient) SendEvent(event string, data interface{}) {
	client.sendEvent(event, data, client.namespace.name)
}

func (client *SocketClient) HasAck() bool {
	return client.ack != nil
}

func (client *SocketClient) GetAck() *Ack {
	return client.ack
}