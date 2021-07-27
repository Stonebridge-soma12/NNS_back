package model

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

type User struct {
	Id           int64          `db:"id"`
	Name         string         `db:"name"`
	ProfileImage sql.NullString `db:"profile_image"`
	Description  sql.NullString `db:"description"`
	LoginId      sql.NullString `db:"login_id"`
	LoginPw      NullBytes      `db:"login_pw"`
	Status       Status         `db:"status"`
	CreateTime   time.Time      `db:"create_time"`
	UpdateTime   time.Time      `db:"update_time"`
}

func NewUser(id string, pw []byte) User {
	return User{
		Name:         "Anonymous",
		ProfileImage: sql.NullString{},
		Description:  sql.NullString{},
		LoginId: sql.NullString{
			String: id,
			Valid:  true,
		},
		LoginPw: NullBytes{
			Bytes: pw,
			Valid: true,
		},
		Status:     StatusEXIST,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
}

// NullBytes represents a JSON that may be null.
// NullBytes implements the Scanner interface so
// it can be used as a scan destination
type NullBytes struct {
	Bytes []byte
	Valid bool
}

// Scan implements the Scanner interface.
func (nb *NullBytes) Scan(value interface{}) error {
	if value == nil {
		nb.Bytes, nb.Valid = nil, false
		return nil
	}
	nb.Valid = true

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprintf("failed to interface conversion"))
	}
	nb.Bytes = bytes
	return nil
}

// Value implements the driver Valuer interface.
func (nb NullBytes) Value() (driver.Value, error) {
	if !nb.Valid {
		return nil, nil
	}
	return nb.Bytes, nil
}

type SelectUserClassifier interface {
	classify(builder *squirrel.SelectBuilder)
}

type selectUserClassifierFunc func(builder *squirrel.SelectBuilder)

func (f selectUserClassifierFunc) classify(builder *squirrel.SelectBuilder) {
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

func SelectUser(db *sqlx.DB, classifier SelectUserClassifier) (User, error) {
	builder := squirrel.Select(
		"u.id",
		"u.name",
		"u.profile_image",
		"u.description",
		"u.login_id",
		"u.login_pw",
		"u.status",
		"u.create_time",
		"u.update_time").
		From("user u").
		Where(squirrel.Eq{"u.status": StatusEXIST})
	classifier.classify(&builder)
	query, args, err := builder.ToSql()
	if err != nil {
		return User{}, errors.Wrap(err, "failed to build sql")
	}

	u := User{}
	err = db.QueryRowx(query, args...).StructScan(&u)
	return u, err
}

func (u User) Insert(db *sqlx.DB) (int64, error) {
	result, err := db.NamedExec(
		`INSERT INTO user
				(name,
				 profile_image,
				 description,
				 login_id,
				 login_pw,
				 status,
				 create_time,
				 update_time)
				VALUES (:name,
						:profile_image,
						:description,
						:login_id,
						:login_pw,
						:status,
						:create_time,
						:update_time);`, u)
	if err != nil {
		return 0, err
	}

	u.Id, err = result.LastInsertId()
	return u.Id, err
}

func (u User) Update(db *sqlx.DB) error {
	u.UpdateTime = time.Now()

	_, err := db.NamedExec(
		`UPDATE user
				SET name          = :name,
					profile_image = :profile_image,
					description   = :description,
					login_id      = :login_id,
					login_pw      = :login_pw,
					status        = :status,
					create_time   = :create_time,
					update_time   = :update_time
				WHERE id = :id
				  AND status = 'EXIST';`, u)

	return err
}

func (u User) Delete(db *sqlx.DB) error {
	u.Status = StatusDELETED
	return u.Update(db)
}
