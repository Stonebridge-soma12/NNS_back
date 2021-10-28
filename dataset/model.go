package dataset

import (
	"database/sql"
	"time"
)

type Dataset struct {
	ID          int64          `db:"id"`      // primary key
	UserID      int64          `db:"user_id"` // uploader ID
	DatasetNo   int64          `db:"dataset_no"`
	URL         string         `db:"url"`         // AWS S3 URL, unique
	Name        sql.NullString `db:"name"`        // dataset name, unique
	Description sql.NullString `db:"description"` // dataset description
	Public      sql.NullBool   `db:"public"`
	Status      string         `db:"status"`
	CreateTime  time.Time      `db:"create_time"`
	UpdateTime  time.Time      `db:"update_time"`
	ImageId     sql.NullInt64  `db:"image_id"` // thumbnail image
	Kind        Kind           `db:"kind"`     // dataset kind

	// additional
	InLibrary    sql.NullBool   `db:"in_library"`
	Usable       sql.NullBool   `db:"usable"`
	ThumbnailUrl sql.NullString `db:"thumbnail_url"`
}

type Kind string

const (
	KindUnknown Kind = "UNKNOWN"
	KindImages  Kind = "IMAGES"
	KindText    Kind = "TEXT"
)

const (
	maxDatasetName        = 100
	maxDatasetDescription = 2000
)

const (
	EXIST    = "EXIST"
	DELETED  = "DELETED"
	UPLOADED = "UPLOADED"
)
