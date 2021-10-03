package train


type TrainRepository interface {
	FindNextTrainNo(userId int64) (int64, error)

	Insert(train Train) (int64, error)
	Delete(opts ...Option) error
	Find(opts ...Option) (Train, error)
	FindAll(opts ...Option) ([]Train, error)
	Update(train Train, opts ...Option) error
}