package queue

import (
	"net"
	"encoding/json"
)

type QueueConfig struct {
	port	int			`json:"port"`
	host    string  	`json:"host"`
}

const(
	protocol = "tcp"
	defaultPort = 4987
	defaultHost = "localhost"
)

func defaultConfig() *QueueConfig {
	return &QueueConfig{
		port: defaultPort,
		host: defaultHost,
	}
}

func parseConfig(src []byte) (*QueueConfig, error) {
	var config QueueConfig
	err := json.Unmarshal(src, &config)
	if err != nil {
		nil, err
	}
	return &config, nil
}

type Queue struct {
	conn 	net.Listener
	conf	*QueueConfig
	out     chan []byte
}

func NewQueueFromJson(src []byte, out <-chan []byte) (queue *Queue, err error) {
	conf, err := parseConfig(src)
	if err != nil{
		return nil, err
	}
	queue = &Queue {
		conf: conf,
	}
	return queue, nil
}

func NewQueue(out <-chan []byte) (*Queue) {
	return &Queue{
		conf: defaultConfig(),
		out: out,
	}
}

func (queue *Queue) Run() (error) {
	// todo assemble url
	conn, err := net.Listen(protocol, "")
	if err != nil {
		return err
	}
	queue.conn = conn
	return nil
}

func (queue *Queue) Stop() {

}