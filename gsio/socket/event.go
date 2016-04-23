package socket

// Different callback types
type EventCallback func(*Client, interface{})
type RoomCreatedCallback func(room *Room)
type ConnectCallback func(*Client)
type DisconnectCallback func(*Client)

type Listenable interface {
	// Listen to event
	ListenCallback(string, EventCallback)
	// Listen to event
	Listen(string, chan<- *Event)
}

type registerListener struct {
	listenerType 		EventType
	event        		string
	eventc       		chan<- *Event
	clientc      		chan<- *Client
	ConnectCallback
	DisconnectCallback
	EventCallback
}

type EventType int

const(
	connectCall EventType = iota
	connectChan
	disconnectCall
	disconnectChan
	customCall
	customChan
)

type Event struct {
	// type of the event
	EventType EventType
	// Name of the event
	Name      string
	// Client receiving event
	Client    *Client
	// event data
	Data      interface{}
}

type Listeners struct {
	// registration channel
	liregc     chan registerListener
	// event listeners
	callbacks  map[string] []EventCallback
	//  connection callbacks
	cCallbacks []ConnectCallback
	// disconnect listeners
	dCallbacks []DisconnectCallback
	// channel event listeners
	eChans     map[string] chan *Event
	// connect event channel
	cChan      chan *Client
	// disconnect event channel
	dChan      chan *Client
}

func newListeners() *Listeners {
	return &Listeners{
		liregc:       	make(chan registerListener),
		callbacks:		make(map[string] []EventCallback),
		cCallbacks: 	make([]ConnectCallback, 0),
		dCallbacks: 	make([]DisconnectCallback, 0),
		eChans: 		make(map[string] chan *Event),
	}
}

func (lst *Listeners) AddConnectCallback(callback ConnectCallback) {
	lst.liregc <- registerListener{
		listenerType: connectCall,
		ConnectCallback: callback,
	}
}

func (lst *Listeners) AddDisconnectCallback(callback DisconnectCallback) {
	lst.liregc <- registerListener{
		listenerType: disconnectCall,
		DisconnectCallback: callback,
	}
}

func (lst *Listeners) ListenCallback(event string, callback EventCallback) {
	lst.liregc <- registerListener{
		listenerType: customCall,
		event: event,
		EventCallback: callback,
	}
}

func (lst *Listeners) SetConnectChan(clientc chan<- *Client) {
	lst.liregc <- registerListener{
		listenerType: connectChan,
		clientc: clientc,
	}
}

func (lst *Listeners) SetDisconnectChan(clientc chan<- *Client) {
	lst.liregc <- registerListener{
		listenerType: disconnectChan,
		clientc: clientc,
	}
}

func (lst *Listeners) Listen(event string, eventc chan<- *Event) {
	lst.liregc <- registerListener{
		listenerType: customChan,
		event: event,
		eventc: eventc,
	}
}