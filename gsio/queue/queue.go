package queue

import (
	"net"
	"encoding/json"
)


type Queue struct {
	conn 	net.Listener
	conf	*Config
	out     chan []byte
}

func NewQueue(conf *Config, out <-chan []byte) (*Queue) {
	return &Queue{
		conf: defaultConfig(),
		out: out,
	}
}

func (queue *Queue) Run() (error) {
	// todo assemble url
	conn, err := net.Listen(Protocol, "")
	if err != nil {
		return err
	}
	queue.conn = conn
	return nil
}

func (queue *Queue) Stop() {

}