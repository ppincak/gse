package socket


type broadcastMessage struct {
	event   string			`json:"event"`
	data 	interface{}		`json:"data"`
}


