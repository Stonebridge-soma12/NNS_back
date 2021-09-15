package trainMonitor

type Monitor struct {
	Epoch    Epoch
	TrainLog TrainLog
	Channel  chan Monitor
}

func (m Monitor) Send() {
	m.Channel <- m
}
