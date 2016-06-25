package socket

import (
	"sync"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/ppincak/gse/socket/transport"
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
					case eventListener:
						logrus.Infof("Namespace: %s - registering listener for event: %s", namespace.name,  msg.event)
						if listeners, ok := l.events[msg.event]; ok {
							l.events[msg.event] = append(listeners, msg.EventListener)
						} else {
							l.events[msg.event] = []EventListener {msg.EventListener}
						}
				}
		case evt := <- namespace.evc:
			switch evt.listenerType {
				case connectListener:
					for _, listener := range l.clientCon {
						listener(evt.client.wrap(namespace))
					}
				case disconnectListener:
					for _, listener := range l.clientDis {
						listener(evt.client.wrap(namespace))
					}
				case eventListener:
					if listeners, ok := l.events[evt.mame]; ok {
						for _, listener := range listeners {
							socketClient := evt.client.wrap(namespace)
							socketClient.ack = evt.ack
							listener(socketClient, evt.data)
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

func (namespace *Namespace) AddRoom(roomName string) *Room {
	namespace.mtx.Lock()
	room := NewRoom(namespace, roomName)
	namespace.rooms[room.uuid] = room
	namespace.mtx.Unlock()
	return room
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

func (namespace *Namespace) GetRooms() []*Room {
	namespace.mtx.RLock()
	defer namespace.mtx.RUnlock()
	rooms := make([]*Room, len(namespace.rooms))
	i := 0
	for _, room := range namespace.rooms {
		rooms[i] = room
		i++
	}
	return rooms
}

func (namespace *Namespace) RemoveRoom(roomName string) {
	namespace.mtx.Lock()
	defer namespace.mtx.Unlock()
	for _, room := range namespace.rooms {
		if room.name == roomName {
			room.Destroy()
			delete(namespace.rooms, roomName)
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

func (namespace *Namespace) GetClients() []*Client {
	namespace.mtx.RUnlock()
	defer namespace.mtx.RUnlock()
	clients := make([]*Client, len(namespace.clients))
	i := 0
	for _, client := range namespace.clients {
		clients[i] = client
		i++
	}
	return clients
}

func (namespace *Namespace) SendEvent(event string, data interface{}) {
	namespace.mtx.RLock()
	for _, client := range namespace.clients {
		client.sendEvent(event, data, namespace.name)
	}
	namespace.mtx.RUnlock()
}

func (namespace *Namespace) addClient(client *Client) {
	namespace.mtx.Lock()
	namespace.clients[client.uuid] = client
	namespace.mtx.Unlock()
	namespace.evc <- &listenerEvent{
		listenerType: connectListener,
		client: client,
	}
	client.notify(transport.Connect, namespace.name)
}

func (namespace *Namespace) removeClient(client *Client) {
	namespace.mtx.Lock()
	delete(namespace.clients, client.uuid)
	namespace.mtx.Unlock()
	namespace.evc <- &listenerEvent{
		listenerType: disconnectListener,
		client: client,
	}
	client.notify(transport.Disconnect, namespace.name)
}