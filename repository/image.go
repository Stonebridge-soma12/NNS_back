package repository

import (
	"nns_back/model"
)

type ImageRepository interface {
	SelectImage(userId, id int64) (model.Image, error)
	Insert(image model.Image) (int64, error)
}
