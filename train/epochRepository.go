package train

import "github.com/elixter/Querybuilder"

type EpochRepository interface {
	Insert(epoch Epoch) error
	Find(opts ...query.Option) (Epoch, error)
	Delete(opts ...query.Option) error
	FindAll(opts ...query.Option) ([]Epoch, error)
}
