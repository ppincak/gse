package socket


type transportmessage struct {
	Event	string 			`json:"event"`
	Data    interface{} 	`json:"message"`
}


