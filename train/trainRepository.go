package train

import "github.com/Masterminds/squirrel"

type TrainRepository interface {
	FindNextTrainNo(userId int64) (int64, error)

	Insert(train Train) (int64, error)
	Delete(opts ...Option) error
	Find(opts ...Option) (Train, error)
	FindAll(opts ...Option) ([]Train, error)
	Update(train Train, opts ...Option) error
}

type SelectTrainOption interface {
	apply(builder *squirrel.SelectBuilder)
}

type selectTrainOptionFunc func(builder *squirrel.SelectBuilder)

func (f selectTrainOptionFunc) apply(builder *squirrel.SelectBuilder) {
	f(builder)
}

func ClassifiedByUserId(userId int64) SelectTrainOption {

}