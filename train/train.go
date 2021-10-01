package train

import (
	"encoding/json"
	"io"
	"net/http"
)

type Train struct {
	Id        int64   `db:"id" json:"id"`
	UserId    int64   `db:"user_id" json:"user_id"`
	TrainNo   int     `db:"train_no" json:"train_no"`
	ProjectId int64   `db:"project_id" json:"project_id"`
	Status    string  `db:"status" json:"status"`
	Acc       float64 `db:"acc" json:"acc"`
	Loss      float64 `db:"loss" json:"loss"`
	ValAcc    float64 `db:"val_acc" json:"val_acc"`
	ValLoss   float64 `db:"val_loss" json:"val_loss"`
	Epochs    int     `db:"epochs" json:"epochs"`
	Name      string  `db:"name" json:"name"`
	Url       string  `db:"url" json:"url"` // saved model url
}

const (
	TrainStatusFinish = "FIN"
	TrainStatusTrain  = "TRAIN"
	TrainStatusError  = "ERR"
	TrainStatusDelete = "DEL"
)

func (t *Train) Bind(r *http.Request) error {
	buffer, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(buffer, t)
	if err != nil {
		return err
	}

	return nil
}

func (t *Train) Update(e Epoch) {
	t.Acc = e.Acc
	t.Loss = e.Loss
	t.ValAcc = e.ValAcc
	t.ValLoss = e.ValLoss
	t.Epochs = e.Epoch
}

type TrainRepository interface {
	Insert(train Train) error
	Delete(opts ...Option) error
	Find(opts ...Option) (Train, error)
	FindAll(opts ...Option) ([]Train, error)
	Update(train Train, opts ...Option) error
}
