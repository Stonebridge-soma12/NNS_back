package model

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type Image struct {
	Id         int64     `db:"id"`
	UserId     int64     `db:"user_id"`
	Url        string    `db:"url"`
	CreateTime time.Time `db:"create_time"`
	UpdateTime time.Time `db:"update_time"`
}

func NewImage(userId int64, url string) Image {
	return Image{
		UserId:     userId,
		Url:        url,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
}

func (i *Image) Insert(db *sqlx.DB) (int64, error) {
	result, err := db.NamedExec(`
INSERT INTO image (user_id, 
                   url, 
                   create_time, 
                   update_time)
VALUES (:user_id,
        :url,
        :create_time,
        :update_time);`, i)
	if err != nil {
		return 0, err
	}

	i.Id, err = result.LastInsertId()
	return i.Id, err
}