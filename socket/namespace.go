package socket

import (
	"sync"
	"errors"
	"github.com/sirupsen/logrus"
	"com.grid/gse/socket/transport"
)

type Namespace struct {
	name      	string
	// reference to server
	server		*Server
	// rooms in the namespace
	rooms		map[string]*Room
	// clients in the namespace
	clients		map[string]*Client
	// listeners
	*Listeners
	// events 		channel
	evc       	chan *listenerEvent
	// stopping channel
	stopc     	chan struct{}
	// lock
	mtx       	*sync.RWMutex
}

func rootNamespace(server *Server) *Namespace {
	return newNamespace("/", server)
}

func newNamespace(name string, server *Server) *Namespace {
	return &Namespace{
		name: 		name,
		server: 	server,
		rooms: 		make(map[string]*Room),
		clients:	make(map[string]*Client),
		Listeners:	newListeners(),
		evc: 		make(chan *listenerEvent, server.conf.EventBufferSize),
		stopc: 		make(chan struct{}),
		mtx:        new(sync.RWMutex),
	}
}

func (namespace *Namespace)  Run() {
	logrus.Infof("Running namespace: %s routine", namespace.name)

	l := namespace.Listeners
	for {
		select {
			case msg := <- l.liregc:
				switch msg.listenerType {
					case connectListener:
						logrus.Infof("Namespace: %s - registering connect listener", namespace.name)
						l.clientCon = append(l.clientCon, msg.ConnectListener)
					case disconnectListener:
						logrus.Infof("Namespace: %s - registering disconnect listener", namespace.name)
						l.clientDis = append(l.clientDis, msg.DisconnectListener)
					case roomAddListener:
						logrus.Infof("Namespace: %s - adding listener to room creation", namespace.name)
						l.roomAdd = append(l.roomAdd, msg.RoomAddListener)
					case roomRemListener:
						logrus.Infof("Namespace: %s - removing listener to room creation", namespace.name)
						l.roomRem = append(l.roomRem, msg.RoomRemListener)
					case eventListener:
						logrus.Infof("Namespace: %s - registering listener for event: %s", namespace.name,  msg.event)
						if listeners, ok := l.events[msg.event]; ok {
							l.events[msg.event] = append(listeners, msg.EventListener)
						} else {
							l.events[msg.event] = []EventListener {msg.EventListener}
						}
				}
		case evt := <- namespace.evc:
			switch evt.ListenerType {
				case connectListener:
					for _, listener := range l.clientCon {
						listener(evt.Client.wrap(namespace))
					}
				case disconnectListener:
					for _, listener := range l.clientDis {
						listener(evt.Client.wrap(namespace))
					}
				case roomAddListener:
					for _, listener := range l.roomAdd {
						listener(evt.Room)
					}
				case roomRemListener:
					for _, listener := range l.roomRem {
						listener(evt.Room)
					}
				case eventListener:
					if listeners, ok := l.events[evt.Name]; ok {
						for _, listener := range listeners {
							listener(evt.Client.wrap(namespace), evt.Data)
						}
					}
				}
			case <- namespace.stopc:
				logrus.Infof("Stopping namespace: %s routine", namespace.name)
				return
		}
	}
}

func (namespace *Namespace) Stop() {
	logrus.Infof("Stopped namespace: %s routine", namespace.name)
	namespace.stopc <- struct{}{}
}

func (namespace *Namespace) GetName() string {
	return namespace.name
}

func (namespace *Namespace) AddRoom(roomName string) string {
	namespace.mtx.Lock()
	room := NewRoom(namespace, roomName)
	namespace.rooms[room.uuid] = room
	namespace.mtx.Unlock()
	namespace.evc <- &listenerEvent{
		ListenerType: roomAddListener,
		Room: room,
	}
	return room.uuid
}

func (namespace *Namespace) GetRoom(roomName string) (*Room, error) {
	namespace.mtx.RLock()
	defer namespace.mtx.RUnlock()
	for _, room := range namespace.rooms {
		if room.name == roomName {
			return room, nil
		}
	}
	return nil, errors.New("Room not found")
}

func (namespace *Namespace) RemoveRoom(roomName string) {
	namespace.mtx.Lock()
	defer namespace.mtx.Unlock()
	for _, room := range namespace.rooms {
		if room.name == roomName {
			room.DestroyRoom()
			delete(namespace.rooms, roomName)
			namespace.evc <- &listenerEvent{
				ListenerType: roomRemListener,
				Room: room,
			}
			return
		}
	}
}

func (namespace *Namespace) joinRoom(roomName string, client *Client) {
	namespace.mtx.RLock()
	defer namespace.mtx.RUnlock()
	for _, room := range namespace.rooms {
		if room.name == roomName {
			room.addClient(client)
			client.joinRoom(room)
			return
		}
	}
}

func (namespace *Namespace) GetClient(sessiondId string) *Client {
	namespace.mtx.RUnlock()
	defer namespace.mtx.RUnlock()
	return namespace.clients[sessiondId]
}

func (namespace *Namespace) SendEvent(event string, data interface{}) {
	namespace.mtx.RLock()
	for _, client := range namespace.clients {
		client.SendEvent(event, data, namespace.name)
	}
	namespace.mtx.RUnlock()
}

func (namespace *Namespace) addClient(client *Client) {
	namespace.mtx.Lock()
	namespace.clients[client.uuid] = client
	namespace.mtx.Unlock()
	namespace.evc <- &listenerEvent{
		ListenerType: connectListener,
		Client: client,
	}

	client.SendPacket(&transport.Packet{
		PacketType: transport.Connect,
		Endpoint: namespace.name,
	})
}

func (namespace *Namespace) removeClient(client *Client) {
	namespace.mtx.Lock()
	delete(namespace.clients, client.uuid)
	namespace.mtx.Unlock()
	namespace.evc <- &listenerEvent{
		ListenerType: disconnectListener,
		Client: client,
	}

	client.SendPacket(&transport.Packet{
		PacketType: transport.Disconnect,
		Endpoint: namespace.name,
	})
}