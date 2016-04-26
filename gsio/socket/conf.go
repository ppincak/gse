package socket

const(
	ServerName 			= "default"
	EventBufferSize 	= 100
	ReadBufferSize 		= 1024
	WriteBufferSize 	= 1024
	MaxNumOfClients 	= 10000
	MaxNumOfRooms 		= 5000
)

type ServerConf struct {
	ServerName      	string	`json:"serverName"`
	EventBufferSize 	int     `json:"eventBuffer"`
	ReadBufferSize  	int 	`json:"readBufferSize"`
	WriteBufferSize 	int 	`json:"writeBufferSize"`
	MaxNumOfClients 	int32 	`json:"numberOfClients"`
	MaxNumOfRooms   	int32	`json:"numberOfRooms"`
}

func DefaultConf() *ServerConf {
	return &ServerConf{
		ServerName: 		ServerName,
		EventBufferSize:	EventBufferSize,
		ReadBufferSize: 	ReadBufferSize,
		WriteBufferSize: 	WriteBufferSize,
		MaxNumOfClients: 	MaxNumOfClients,
		MaxNumOfRooms: 		MaxNumOfRooms,
	}
}