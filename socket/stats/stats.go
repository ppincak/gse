package stats

import "sync/atomic"

const(
	OpenedConnections = iota
	ClosedConnections
	OpenedRooms
	ClosedRooms
	ConnectionFailures
	PacketFailures
)

type Stats struct {
	OpenedConnections  uint64		`json:"openedConnections"`
	ClosedConnections  uint64		`json:"closedConnections"`
	OpenedRooms        uint64		`json:"openedRooms"`
	ClosedRooms        uint64		`json:"closedRooms"`
	ConnectionFailures uint64		`json:"connectionFailures"`
	PacketFailures     uint64		`json:"PacketFailures"`
}

func NewStats() *Stats {
	return &Stats{}
}

func (stats *Stats) Inc(field int) {
	switch field {
		case OpenedConnections:
			atomic.AddUint64(&stats.OpenedConnections, 1)
		case ClosedConnections:
			atomic.AddUint64(&stats.ClosedConnections, 1)
		case OpenedRooms:
			atomic.AddUint64(&stats.OpenedRooms, 1)
		case ClosedRooms:
			atomic.AddUint64(&stats.ClosedRooms, 1)
		case ConnectionFailures:
			atomic.AddUint64(&stats.ConnectionFailures, 1)
		case PacketFailures:
			atomic.AddUint64(&stats.PacketFailures, 1)
	}
}