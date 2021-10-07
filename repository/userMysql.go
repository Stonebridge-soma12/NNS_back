package repository

import (
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"nns_back/model"
	"nns_back/util"
	"time"
)

type userRepositoryImpl struct {
	db *sqlx.DB
}

func NewUserMysqlRepository(db *sqlx.DB) UserRepository {
	return &userRepositoryImpl{
		db: db,
	}
}

func (r *userRepositoryImpl) SelectUser(classifier SelectUserClassifier) (model.User, error) {
	builder := squirrel.Select(
		"u.id",
		"u.name",
		"u.profile_image",
		"u.description",
		"u.email",
		"u.web_site",
		"u.login_id",
		"u.login_pw",
		"u.status",
		"u.create_time",
		"u.update_time").
		From("user u").
		Where(squirrel.Eq{"u.status": util.StatusEXIST})
	classifier.userClassify(&builder)
	query, args, err := builder.ToSql()
	if err != nil {
		return model.User{}, errors.Wrap(err, "failed to build sql")
	}

	user := model.User{}
	err = r.db.QueryRowx(query, args...).StructScan(&user)
	return user, err
}

func (r *userRepositoryImpl) Insert(user model.User) (int64, error) {
	result, err := r.db.NamedExec(
		`INSERT INTO user
				(name,
				 profile_image,
				 description,
				 email,
				 web_site,
				 login_id,
				 login_pw,
				 status,
				 create_time,
				 update_time)
				VALUES (:name,
						:profile_image,
						:description,
				        :email,
				        :web_site,
						:login_id,
						:login_pw,
						:status,
						:create_time,
						:update_time);`, user)
	if err != nil {
		return 0, err
	}

	user.Id, err = result.LastInsertId()
	return user.Id, err
}

func (r *userRepositoryImpl) Update(user model.User) error {
	user.UpdateTime = time.Now()

	_, err := r.db.NamedExec(
		`UPDATE user
				SET name          = :name,
					profile_image = :profile_image,
					description   = :description,
				    email		  = :email,
				    web_site 	  = :web_site,
					login_id      = :login_id,
					login_pw      = :login_pw,
					status        = :status,
					create_time   = :create_time,
					update_time   = :update_time
				WHERE id = :id
				  AND status = 'EXIST';`, user)

	return err
}

func (r *userRepositoryImpl) Delete(user model.User) error {
	user.Status = util.StatusDELETED
	return r.Update(user)
}
