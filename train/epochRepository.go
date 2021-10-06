package train

import "nns_back/query"

type EpochRepository interface {
	Insert(epoch Epoch) error
	Find(opts ...query.Option) (Epoch, error)
	Delete(opts ...query.Option) error
	FindAll(opts ...query.Option) ([]Epoch, error)
}
