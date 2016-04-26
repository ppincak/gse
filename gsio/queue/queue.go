package queue

import (

	"golang.org/x/net/context"
	"com.grid/chsen/gsio/socket"
	//"fmt"
	"time"
	"fmt"
)


type Queue struct {
	queuec      chan interface{}
	stopc		chan struct{}
	server 		*socket.Server
}

func NewQueue() (*Queue) {
	return &Queue{
	}
}

type queueMessage struct {
	context 	context.Context
	req 		interface{}
}

func (queue *Queue) Run() {
	for {
		select {
			case msg := <- queue.queuec:
				switch msg.(type) {
					case *AddRoomReq:
						req := msg.(*AddRoomReq)
						queue.server.AddRoom(req.RoomName)
				}
		}
	}
}

func (queue *Queue) Stop() {

}

func (queue *Queue) AddRoom(ctx context.Context, req *AddRoomReq) (*Response, error) {
	fmt.Println(time.Now())
	return new(Response), nil
}

func (queue *Queue) RemoveRoom(ctx context.Context, req *RemRoomReq) (*Response, error) {
	return new(Response), nil
}