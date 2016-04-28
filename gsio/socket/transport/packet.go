package transport

import "encoding/json"

type PacketType int

const(
	Connect PacketType = iota
	Disconnect
	Event
	Ack
	Error
	BinaryEvent
	BinaryAck
)

var EventMap = map[string] PacketType {
	"connect": 		Connect,
	"disconnect": 	Disconnect,
	"event": 		Event,
	"ack": 			Ack,
	"error": 		Error,
	"binaryack": 	BinaryAck,
	"binaryevent": 	BinaryEvent,
}

type Packet struct {
	// type of packet
	PacketType		PacketType		`json:"type"`
	// namespace
	Endpoint		string			`json:"endpoint"`
	//query string
	Qs				string			`json:"qs"`
	// data
	Data            []interface{}	`json:"data"`
	// id
	Id 				int64			`json:"id"`
	// event name
	Name            string			`json:"name"`
	// event arguments
	Args			[]interface{}	`json:"args"`
}

func Encode(packet *Packet) ([]byte, error) {
	bytes, err := json.Marshal(packet)
	if err != nil {
		return nil, err
	}
	return bytes, err
}

func Decode(bytes []byte) (*Packet, error) {
	var packet Packet
	err := json.Unmarshal(bytes, &packet)
	if err != nil {
		return nil, err
	}
	return &packet, nil
}