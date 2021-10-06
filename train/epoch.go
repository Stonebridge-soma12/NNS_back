package train

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type Epoch struct {
	Id           int64     `db:"id" json:"id"`
	TrainId      int64     `db:"train_id" json:"train_id" header:"train_id"`
	Acc          float64   `db:"acc" json:"accuracy"`
	Epoch        int       `db:"epoch" json:"epoch"`
	Loss         float64   `db:"loss" json:"loss"`
	ValAcc       float64   `db:"val_acc" json:"val_accuracy"`
	ValLoss      float64   `db:"val_loss" json:"val_loss"`
	LearningRate float64   `db:"learning_rate" json:"lr"`
	CreateTime   time.Time `db:"create_time" json:"create_time"`
	UpdateTime   time.Time `db:"update_time" json:"update_time"`
}

func (e *Epoch) Bind(r *http.Request) error {
	buffer, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buffer, e)
	if err != nil {
		return err
	}

	return nil
}
