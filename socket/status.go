package socket

type RoomStatus struct {
	RoomName			string 		`json:"roomName"`
	NumberOfClients     int			`json:"numberOfClients"`
}

type Status struct {
	ServerName  		string			`json:"serverName"`
	NumberOfClients		int				`json:"numberOfClients"`
	NumberOfRooms		int				`json:"numberOfRooms"`
	RoomsStatus			[]RoomStatus	`json:"roomStatus"`
}