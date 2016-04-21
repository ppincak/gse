package queue

import (
	"net"
	"encoding/json"
	"strings"
)


type Queue struct {
	conn 	net.Listener
	conf	*QueueConf
	out     chan []byte
}

func NewQueue(conf *QueueConf, out <-chan []byte) (*Queue) {
	return &Queue{
		conf: defaultConfig(),
		out: out,
	}
}

func (queue *Queue) Run() (error) {
	// todo assemble url
	conn, err := net.Listen(Protocol, queue.assembleUrl())
	if err != nil {
		return err
	}



    queue.conn = conn
	return nil
}

func (queue *Queue) Stop() {

}

func (queue *Queue) assembleUrl() string {
	return strings.Join([]string{queue.conf.Host, ":", queue.conf.Port}, "")
}