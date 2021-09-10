package trainMonitor

import (
	"encoding/json"
	"net/http"
)

type TrainLog struct {
	Id      int    `db:"id" json:"id"`
	TrainId string `db:"train_id" json:"train_id" header:"train_id"`
	Message string `db:"msg" json:"msg"`
}

func (l *TrainLog) Bind(r *http.Request) error {
	var log []byte
	_, err := r.Body.Read(log)
	if err != nil {
		return err
	}

	var res Epoch
	err = json.Unmarshal(log, &res)
	if err != nil {
		return err
	}

	return nil
}

type TrainLogRepository interface {
	Insert(log TrainLog) error
	Delete(opts ...Option) error
	Update(train Train, opts ...Option) error
	Find(opts ...Option) (TrainLog, error)
	FindAll(opts ...Option) ([]TrainLog, error)
}