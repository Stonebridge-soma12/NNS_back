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
	Description string        `db:"description"`
	Config      util.NullJson `db:"config"`
	Content     util.NullJson `db:"content"`
	Status      util.Status   `db:"status"`
	CreateTime  time.Time     `db:"create_time"`
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
		"optimizer":     "adam",
		"learning_rate": 0.001,
		"loss":          "sparse_categorical_crossentropy",
		"metrics":       []interface{}{"accuracy"},
		"batch_size":    32,
		"epochs":        10,
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