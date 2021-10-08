package model

import (
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