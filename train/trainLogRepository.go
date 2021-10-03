package train

type TrainLogRepository interface {
	Insert(log TrainLog) error
	Delete(opts ...Option) error
	Find(opts ...Option) (TrainLog, error)
	FindAll(opts ...Option) ([]TrainLog, error)
}