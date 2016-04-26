package chess

import "com.grid/chsen/gsio/socket"

type Game struct {
	server				*socket.Server
	conc        		chan *socket.Client
	disc        		chan *socket.Client
	comc        		chan *socket.Event
	movec				chan *socket.Event
	joinc				chan *socket.Event

	matches				map[string] *Match
	clientMatch 		map[string] string
}

func (game *Game) register() {
	game.server.AddConnectListener(game.conc)
	game.server.AddDisconnectListener(game.disc)
	game.server.AddRoomAddListener(nil)
	game.server.AddRoomRemoveListener(nil)
	game.server.Listen("join", game.movec)
	game.server.Listen("move", game.joinc)
	game.server.Listen("comment", game.conc)
}

func (game *Game) Run() {
	for {
		select {
			case evt := <- game.joinc:
				if evt.Data != nil {
					m := evt.Data.(map[string] string)
					m["roomName"] = 10
				}
			case evt := <- game.movec:
				if _, ok := game.clientMatch[evt.Client.GetSessionId()]; ok{

				}
			case evt := <- game.comc:
				if evt {

				}
			/*	case client := <- game.conc:

			case client := <- game.disc:*/
		}
	}
}