package repository

import (
	"github.com/jmoiron/sqlx"
	"nns_back/model"
)

type imageMysqlRepository struct {
	db *sqlx.DB
}

func NewImageMysqlRepository(db *sqlx.DB) ImageRepository {
	return &imageMysqlRepository{
		db: db,
	}
}

func (r *imageMysqlRepository) SelectImage(userId, id int64) (model.Image, error) {
	image := model.Image{}
	err := r.db.QueryRowx(`
SELECT i.id, 
       i.user_id, 
       i.url, 
       i.create_time, 
       i.update_time
FROM image i
WHERE i.id = ?
  AND i.user_id = ?;`, id, userId).StructScan(&image)
	return image, err
}

func (r *imageMysqlRepository) Insert(image model.Image) (int64, error) {
	result, err := r.db.NamedExec(`
INSERT INTO image (user_id, 
                   url, 
                   create_time, 
                   update_time)
VALUES (:user_id,
        :url,
        :create_time,
        :update_time);`, image)
	if err != nil {
		return 0, err
	}

	image.Id, err = result.LastInsertId()
	return image.Id, err
}
