package monitor


type Monitor struct {
	stats		*Stats
	status  	*Status
}

func NewMonitor() *Monitor {
	return &Monitor{

	}
}

/**

func (server *Server) getStatus() (Status) {
	server.mtx.RLock()
	defer server.mtx.RUnlock()
	numRooms := len(server.rooms)
	roomStatus := make([]RoomStatus, numRooms)

	i := 0
	for _, room := range server.rooms {
		roomStatus[i] = RoomStatus{
			RoomName: room.name,
			NumberOfClients: len(room.clients),
		}
		i++
	}
	server.mtx.RUnlock()

	return Status{
		ServerName: server.conf.ServerName,
		NumberOfClients: len(server.clients),
		NumberOfRooms: numRooms,
		RoomsStatus: roomStatus,
	}
}

func (server *Server) getStats() (monitor.Stats) {
	return *server.stats
}


*/