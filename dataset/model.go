package dataset

import (
	"database/sql"
	"time"
)

type Dataset struct {
	ID          int64          `db:"id"`          // primary key
	UserID      int64          `db:"user_id"`     // uploader ID
	URL         string         `db:"url"`         // AWS S3 URL, unique
	Name        sql.NullString `db:"name"`        // dataset name, unique
	Description sql.NullString `db:"description"` // dataset description
	CreateTime  time.Time      `db:"create_time"`
	UpdateTime  time.Time      `db:"update_time"`
}

const (
	maxDatasetName        = 100
	maxDatasetDescription = 2000
)
