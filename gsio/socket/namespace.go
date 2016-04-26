package socket

import (
	"sync"
	"errors"
	"github.com/sirupsen/logrus"
)

type Namespace struct {
	name			string
	// reference to server
	server			*Server
	// rooms in the namespace
	rooms      		map[string]*Room
	// clients in the namespace
	clients    		map[string]*Client
	//
	listeners 		*Listeners
	// events 		channel
	evc          	chan *Event
	//
	stopc			chan struct{}
	//
	mtx				*sync.RWMutex
}

func rootNamespace(server *Server) *Namespace {
	return newNamespace("/", nil)
}

func newNamespace(name string, server *Server) *Namespace {
	return &Namespace{
		name: 		name,
		server: 	server,
		rooms: 		make(map[string]*Room),
		clients:	make(map[string]*Client),
		evc: 		make(chan *Event, server.conf.EventBufferSize),
	}
}

func (namespace *Namespace)  Run() {
	logrus.Info("Running namespace routine")

	l := namespace.listeners
	for {
		select {
			case msg := <- l.liregc:
				switch msg.listenerType {
					case connectListener:
						logrus.Infof("Registering connect listener")
						l.clientCon = append(l.clientCon, msg.ConnectListener)
					case disconnectListener:
						logrus.Infof("Registering disconnect listener")
						l.clientDis = append(l.clientDis, msg.DisconnectListener)
					case roomAddListener:
						logrus.Infof("Adding listener to room creation")
						l.roomAdd = append(l.roomAdd, msg.RoomAddListener)
					case roomRemListener:
						logrus.Infof("Removing listener to room creation")
						l.roomRem = append(l.roomRem, msg.RoomRemListener)
					case eventListener:
						logrus.Infof("Registering listener for event: %s", msg.event)
						if listeners, ok := l.events[msg.event]; ok {
							l.events[msg.event] = append(listeners, msg.EventListener)
						} else {
							l.events[msg.event] = []EventListener {msg.EventListener}
						}
				}
		case evt := <- namespace.evc:
			switch evt.EventType {
				case connectListener:
					for _, listener := range l.clientCon {
						listener(evt.Client)
					}
				case disconnectListener:
					for _, listener := range l.clientDis {
						listener(evt.Client)
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
							listener(evt.Client, evt.Data)
						}
					}
				}
			case <- namespace.stopc:
				logrus.Info("Namespace routine stopped")
				return
		}
	}
}

func (namespace *Namespace) Stop() {
	logrus.Info("Stopping socket server worker")
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
	namespace.statIncr(OpenedRooms)
	namespace.evc <- &Event{
		EventType: roomAddListener,
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
			namespace.statIncr(ClosedRooms)
			namespace.evc <- &Event{
				EventType: roomRemListener,
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
	return nil
}

func (namespace *Namespace) addClient(client *Client) {
	namespace.mtx.Lock()
	namespace.clients[client.uuid] = client
	namespace.mtx.Unlock()
	namespace.statIncr(OpenedClients)
	namespace.evc <- &Event{
		EventType: connectListener,
		Client: client,
	}
}

func (namespace *Namespace) removeClient(client *Client) {
	namespace.mtx.Lock()
	delete(namespace.clients, client.uuid)
	namespace.mtx.Unlock()
	namespace.statIncr(ClosedClients)
	namespace.evc <- &Event{
		EventType: disconnectListener,
		Client: client,
	}
}

func (namespace *Namespace) statIncr(stat int) {
	namespace.server.stats.Inc(stat)
}