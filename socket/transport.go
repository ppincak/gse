package socket


// Message used to transport data
type TransportMessage struct {
	Event		string 			`json:"event"`
	Data    	interface{} 	`json:"data"`
}