package socket

type EventCallback func(*Client, interface{})

type ConnectCallback func(*Client)

type DisconnectCallback func(*Client)

type Listenable interface {

	Listen(string, EventCallback)
}

type listenerMessage struct {
	event 		string
	callback	EventCallback
}

type event struct {
	name 		string
	client  	*Client
	data 		interface{}
}