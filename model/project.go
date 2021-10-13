package model

import (
	"database/sql"
	"encoding/json"
	"nns_back/util"
	"time"
)

type Project struct {
	Id          int64          `db:"id"`
	ShareKey    sql.NullString `db:"share_key"`
	UserId      int64          `db:"user_id"`
	ProjectNo   int            `db:"project_no"`
	Name        string         `db:"name"`
	Description string         `db:"description"`
	Config      util.NullJson  `db:"config"`
	Content     util.NullJson  `db:"content"`
	Status      util.Status    `db:"status"`
	CreateTime  time.Time      `db:"create_time"`
	UpdateTime  time.Time      `db:"update_time"`
}

func NewProject(userId int64, projectNo int, name, description string) Project {
	return Project{
		ShareKey:    sql.NullString{Valid: false},
		UserId:      userId,
		ProjectNo:   projectNo,
		Name:        name,
		Description: description,
		Config:      DefaultConfig(),
		Content:     DefaultContent(),
		Status:      util.StatusEXIST,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
}

func DefaultConfig() util.NullJson {
	defaultValue := map[string]interface{}{
		"optimizer_name": "Adam",
		"optimizer_config": map[string]interface{}{
			"learning_rate": 0.001,
			"beta_1":        0.9,
			"beta_2":        0.000,
			"epsilon":       1e-07,
			"amsgrad":       false,
		},
		"loss":       "binary_crossentropy",
		"metrics":    []interface{}{"accuracy"},
		"batch_size": 32,
		"epochs":     10,
		"early_stop": map[string]interface{}{
			"usage":    true,
			"monitor":  "loss",
			"patience": 2,
		},
		"learning_rate_reduction": map[string]interface{}{
			"usage":    true,
			"monitor":  "val_accuracy",
			"patience": 2,
			"factor":   0.25,
			"min_lr":   0.0000003,
		},
	}
	defaultBytes, _ := json.Marshal(defaultValue)
	return util.NullJson{
		Json:  json.RawMessage(defaultBytes),
		Valid: true,
	}
}

func DefaultContent() util.NullJson {
	defaultValue := map[string]interface{}{
		"output": "",
		"layers": []interface{}{},
	}
	defaultBytes, _ := json.Marshal(defaultValue)
	return util.NullJson{
		Json:  json.RawMessage(defaultBytes),
		Valid: true,
	}
}
