package repository

import (
	"github.com/Masterminds/squirrel"
	"nns_back/model"
)

type UserRepository interface {
	SelectUser(classifier SelectUserClassifier) (model.User, error)
	Insert(user model.User) (int64, error)
	Update(user model.User) error
	Delete(user model.User) error
}

type SelectUserClassifier interface {
	userClassify(builder *squirrel.SelectBuilder)
}

type selectUserClassifierFunc func(builder *squirrel.SelectBuilder)

func (f selectUserClassifierFunc) userClassify(builder *squirrel.SelectBuilder) {
	f(builder)
}

func ClassifiedById(userId int64) SelectUserClassifier {
	return selectUserClassifierFunc(func(builder *squirrel.SelectBuilder) {
		*builder = builder.Where(squirrel.Eq{"u.id": userId})
	})
}

func ClassifiedByLoginId(loginId string) SelectUserClassifier {
	return selectUserClassifierFunc(func(builder *squirrel.SelectBuilder) {
		*builder = builder.Where(squirrel.Eq{"u.login_id": loginId})
	})
}
