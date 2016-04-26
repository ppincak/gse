package monitor

type Status struct {
	ServerName  		string				`json:"serverName"`
	NumberOfClients		uint32				`json:"numberOfClients"`
	NumberOfRooms		uint32				`json:"numberOfRooms"`
	NumberOfNamespaces	uint32				`json:"numberofNamespaces"`
	Namespaces			[]NamespaceStatus	`json:"namespaces"`
}

type NamespaceStatus struct {
	Name				string  `json:"name"`
	NumberOfRooms		uint32 	`json:"numberOfRooms"`
	NumberOfClients		uint32	`json:"numberOfClients"`
}

const (
	NumberOfClients int = iota
	NumberOfRooms
)