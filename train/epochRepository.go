package train

type EpochRepository interface {
	Insert(epoch Epoch) error
	Find(opts ...Option) (Epoch, error)
	Delete(opts ...Option) error
	FindAll(opts ...Option) ([]Epoch, error)
}
