package socket

// Callbacks
type EventCallback func(*Client, interface{})

type ConnectCallback func(*Client)

type DisconnectCallback func(*Client)

type Listenable interface {
	// Listen to event
	Listen(string, EventCallback)
}

type registerListener struct {
	listenerType	EventType
	event 			string
	ConnectCallback
	DisconnectCallback
	EventCallback
}

type EventType int

const(
	Connect EventType = iota
	Disconnect
	Custom
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