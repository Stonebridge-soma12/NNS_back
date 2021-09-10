package trainMonitor

type Monitor struct {
	Epoch    Epoch
	TrainLog TrainLog
}

func (m Monitor) Send() {
	ch := make(chan Monitor)

	ch <- m
}