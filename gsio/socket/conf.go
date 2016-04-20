package socket

const(
	ServerName = "default"
	ServerHost = "localhost"
	ServerPort = 9864
	MaxNumOfClients = 10000
	MaxNumOfRooms = 5000
)

type Conf struct {
	ServerName  		string	`json:"serverName"`
	ServerHost			string	`json:"serverHost"`
	ServerPort        	int		`json:"serverPort"`
	MaxNumOfClients 	int32 	`json:"numberOfClients"`
	MaxNumOfRooms   	int32	`json:"numberOfRooms"`
}

func DefaultConf() *Conf {
	return &Conf{
		ServerName: ServerName,
		ServerHost: ServerHost,
		ServerPort: ServerPort,
		MaxNumOfClients: MaxNumOfClients,
		MaxNumOfRooms: MaxNumOfRooms,
	}
}