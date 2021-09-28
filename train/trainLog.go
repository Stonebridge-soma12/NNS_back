package train

import (
	"encoding/json"
	"io"
	"net/http"
)

type TrainLog struct {
	Id      int    `db:"id" json:"id"`
	TrainId int64  `db:"train_id" json:"train_id" header:"train_id"`
	Message string `db:"msg" json:"msg"`
	Status  int    `db:"status" json:"status"`
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

type TrainLogRepository interface {
	Insert(log TrainLog) error
	Delete(opts ...Option) error
	Find(opts ...Option) (TrainLog, error)
	FindAll(opts ...Option) ([]TrainLog, error)
}
