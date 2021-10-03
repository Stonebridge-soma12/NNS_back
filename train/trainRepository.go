package train

type TrainRepository interface {
	Insert(train Train) error
	Delete(opts ...Option) error
	Find(opts ...Option) (Train, error)
	FindAll(opts ...Option) ([]Train, error)
	Update(train Train, opts ...Option) error
}
