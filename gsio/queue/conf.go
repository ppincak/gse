package queue


type QueueConf struct {
	Host 	string  	`json:"host"`
	Port 	int			`json:"port"`
}

const(
	Protocol = "tcp"
	DefaultPort = 4987
	DefaultHost = "localhost"
)

func defaultConfig() *QueueConf {
	return &QueueConf{
		Port: DefaultPort,
		Host: DefaultHost,
	}
}