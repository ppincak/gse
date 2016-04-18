package socket


type managamentEvent uint8

const (
	addClient 		managamentEvent = 1
	removeClient 	managamentEvent = 2
	addRoom 		managamentEvent = 3
	removeRoom		managamentEvent = 4
)

type management struct {
	msgType managamentEvent
	client  *Client
	room    *Room
}