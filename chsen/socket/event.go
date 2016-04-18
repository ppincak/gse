package socket

type EventCallback func(*Client, string)

type Listenable interface {

	Listen(string, EventCallback)
}

type event struct {
	name 		string
	client  	*Client
	data 		string
}