package stats

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
	incrc              chan int
	statc              chan <-chan Stats
	stopc              chan struct{}
}

const (
	StatsBufferSize = 100
)

func NewStats() *Stats {
	return &Stats{
		incrc: make(chan int, StatsBufferSize),
		statc: make(chan chan Stats),
		stopc: make(chan struct{}),
	}
}

func (stats *Stats) Run() {
	go func(stats *Stats) {
		for {
			select {
				case field := <- stats.incrc:
					switch field {
						case OpenedConnections:
							stats.OpenedConnections++
						case ClosedConnections:
							stats.ClosedConnections++
						case OpenedRooms:
							stats.OpenedRooms++
						case ClosedRooms:
							stats.ClosedRooms++
						case ConnectionFailures:
							stats.ConnectionFailures++
						case PacketFailures:
							stats.PacketFailures++
					}
				case c := <- stats.statc:
					c <- stats.Clone()
				case <- stats.stopc:
					return
			}
		}
	}(stats)
}

func (stats *Stats) Stop() {
	stats.stopc <- struct{}{}
}

func (stats *Stats) Get(c chan Stats) {
	stats.statc <- c
}

func (stats *Stats) Inc(field int) {
	stats.incrc <- field
}

func (stats *Stats) Clone() Stats {
	return Stats {
		OpenedConnections:	stats.OpenedConnections,
		ClosedConnections: 	stats.ClosedConnections,
		OpenedRooms: 		stats.OpenedRooms,
		ClosedRooms: 		stats.ClosedRooms,
		ConnectionFailures: stats.ConnectionFailures,
		PacketFailures:		stats.PacketFailures,
	}
}