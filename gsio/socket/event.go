package socket

// Callbacks
type EventCallback func(*Client, interface{})

type ConnectCallback func(*Client)

type DisconnectCallback func(*Client)

type Listenable interface {
	// Event listening
	Listen(string, EventCallback)
}

type registerListener struct {
	event 		string
	callback	EventCallback
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