package monitor

import (


	"sync/atomic"
)

const (
	NumberOfNamespaces int = iota
	NumberOfClients
	NumberOfRooms
)

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

func (status *Status) Inc(field int) {
	switch field {
		case NumberOfNamespaces:
			atomic.AddUint32(&status.NumberOfNamespaces, 1)
		case NumberOfClients:
			atomic.AddUint32(&status.NumberOfClients, 1)
		case NumberOfRooms:
			atomic.AddUint32(&status.NumberOfRooms, 1)
	}
}