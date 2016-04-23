package socket


// Message used to transport data
type Transportmessage struct {
	Event	string 			`json:"event"`
	Data    interface{} 	`json:"data"`
}