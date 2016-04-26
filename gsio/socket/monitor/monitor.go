package monitor


type Monitor struct {
	stats		*Stats
	status  	*Status
}

func NewMonitor() *Monitor {
	return &Monitor{

	}
}