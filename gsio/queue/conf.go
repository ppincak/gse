package queue


type Config struct {
	Host 	string  	`json:"host"`
	Port 	int			`json:"port"`
}

const(
	Protocol = "tcp"
	DefaultPort = 4987
	DefaultHost = "localhost"
)

func defaultConfig() *Config {
	return &Config{
		Port: DefaultPort,
		Host: DefaultHost,
	}
}