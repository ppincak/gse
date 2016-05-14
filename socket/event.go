package socket

type EventListener func(*SocketClient, interface{})
type RoomAddListener func(room *Room)
type RoomRemListener func(room *Room)
type ConnectListener func(*SocketClient)
type DisconnectListener func(*SocketClient)

type Listenable interface {
	Listen(string, chan<- *listenerEvent)
}

type registerListener struct {
	listenerType 		listenerType
	event        		string
	EventListener
	RoomAddListener
	RoomRemListener
	ConnectListener
	DisconnectListener
}

type listenerType int

const(
	connectListener listenerType = iota
	disconnectListener
	eventListener
)

type listenerEvent struct {
	// type of the event
	ListenerType listenerType
	// Name of the event
	Name         string
	// Client receiving event
	Client       *Client
	// Room
	Room         *Room
	// event data
	Data         interface{}
}

type Listeners struct {
	// registration channel
	liregc 			chan registerListener
	// connect event channel
	clientCon		[]ConnectListener
	// disconnect event channel
	clientDis 		[]DisconnectListener
	// event listeners
	events 			map[string] []EventListener
}

func newListeners() *Listeners {
	return &Listeners{
		liregc:     make(chan registerListener),
		clientCon:  make([]ConnectListener, 0),
		clientDis:	make([]DisconnectListener, 0),
		events: 	make(map[string] []EventListener),
	}
}

func (lst *Listeners) AddConnectListener(listener ConnectListener) {
	lst.liregc <- registerListener{
		listenerType: connectListener,
		ConnectListener: listener,
	}
}

func (lst *Listeners) AddDisconnectListener(listener DisconnectListener) {
	lst.liregc <- registerListener{
		listenerType: disconnectListener,
		DisconnectListener: listener,
	}
}

func (lst *Listeners) Listen(event string, listener EventListener) {
	lst.liregc <- registerListener{
		listenerType: eventListener,
		event: event,
		EventListener: listener,
	}
}