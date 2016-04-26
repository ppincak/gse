package monitor

import "sync/atomic"

const(
	OpenedClients = iota
	ClosedClients
	OpenedRooms
	ClosedRooms
	FailedConnections
	FailedMessages
)

type Stats struct {
	OpenedClients 		uint64	`json:"openedClients"`
	ClosedClients 		uint64	`json:"closedClients"`
	OpenedRooms 		uint64	`json:"openedRooms"`
	ClosedRooms			uint64	`json:"closedRooms"`
	FailedConnections	uint64	`json:"failedConnection"`
	FailedMessages		uint64	`json:"failedMessages"`
}

func NewStats() *Stats {
	stats := new(Stats)
	return stats
}

func (stats *Stats) Inc(field int) {
	switch field {
		case OpenedClients:
			atomic.AddUint64(&stats.OpenedClients, 1)
		case ClosedClients:
			atomic.AddUint64(&stats.ClosedClients, 1)
		case OpenedRooms:
			atomic.AddUint64(&stats.OpenedRooms, 1)
		case ClosedRooms:
			atomic.AddUint64(&stats.ClosedRooms, 1)
		case FailedConnections:
			atomic.AddUint64(&stats.FailedConnections, 1)
		case FailedMessages:
			atomic.AddUint64(&stats.FailedMessages, 1)
	}
}