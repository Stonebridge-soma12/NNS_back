package trainMonitor

import (
	"encoding/json"
	"net/http"
)

const (
	defaultSelectEpochQuery = "select * from Epoch "
	defaultDeleteEpochQuery = "delete from Epoch "
)

type Epoch struct {
	TrainId      string  `db:"train_id" json:"train_id" header:"train_id"`
	Acc          float64 `db:"acc" json:"acc"`
	Epoch        int     `db:"epoch" json:"epoch"`
	Loss         float64 `db:"loss" json:"loss"`
	ValAcc       float64 `db:"val_acc" json:"val_acc"`
	ValLoss      float64 `db:"val_loss" json:"val_loss"`
	LearningRate float64 `db:"learning_rate" json:"lr"`
}

func (e *Epoch) Bind(r *http.Request) error {
	var epoch []byte
	_, err := r.Body.Read(epoch)
	if err != nil {
		return err
	}

	var res Epoch
	err = json.Unmarshal(epoch, &res)
	if err != nil {
		return err
	}

	return nil
}

type EpochRepository interface {
	Insert(epoch Epoch) error
	Find(opts ...Option) (Epoch, error)
	Delete(opts ...Option) error
	FindAll(opts ...Option) ([]Epoch, error)
}
