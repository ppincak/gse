package conf

type Conf struct {
	ServerName  		string	`json:"serverName"`
	ServerHost			string	`json:"serverHost"`
	ServerPort        	int		`json:"serverPort"`
	MaxNumOfClients 	int32 	`json:"numberOfClients"`
	MaxNumOfRooms   	int32	`json:"numberOfRooms"`
}

func NewConf() *Conf {
	return &Conf{

	}
}