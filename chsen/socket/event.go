package socket

type EventCallback func(*Client, interface{})

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