package datasetConfig

import (
	"database/sql"
	"nns_back/util"
	"time"
)

type DatasetConfig struct {
	ID                  int64          `db:"id"`
	ProjectId           int64          `db:"project_id"`
	DatasetId           int64          `db:"dataset_id"`
	Name                string         `db:"name"`
	Shuffle             bool           `db:"shuffle"`
	NormalizationMethod sql.NullString `db:"normalization_method"`
	Label               string         `db:"label"`
	Status              util.Status    `db:"status"`
	CreateTime          time.Time      `db:"create_time"`
	UpdateTime          time.Time      `db:"update_time"`
}
