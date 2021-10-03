package train

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type TrainLog struct {
	Id         int    `db:"id" json:"id"`
	TrainId    int64  `db:"train_id" json:"train_id" header:"train_id"`
	Message    string `db:"msg" json:"msg"`
	StatusCode int    `db:"status_code" json:"status_code"`
	CreateTime time.Time `db:"create_time" json:"create_time"`
	UpdateTime time.Time `db:"update_time" json:"update_time"`
}

func (l *TrainLog) Bind(r *http.Request) error {
	log, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(log, l)
	if err != nil {
		return err
	}

	return nil
}


