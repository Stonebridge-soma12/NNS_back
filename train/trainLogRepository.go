package train

import "nns_back/query"

type TrainLogRepository interface {
	Insert(log TrainLog) error
	Delete(opts ...query.Option) error
	Find(opts ...query.Option) (TrainLog, error)
	FindAll(opts ...query.Option) ([]TrainLog, error)
}