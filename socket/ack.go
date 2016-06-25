package socket

import "github.com/ppincak/gse/socket/transport"

type Ack struct {
	id        	int64
	client		*Client
	namespace   *Namespace
}

func (ack *Ack) SendData(data interface{}) {
	if ack.client.isOpen() {
		ack.client.SendPacket(&transport.Packet{
			PacketType: transport.Ack,
			Endpoint: 	ack.namespace.name,
			Id: 		ack.id,
			Data: 		data,
		})
	}
}