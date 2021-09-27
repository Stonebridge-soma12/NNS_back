package trainMonitor

import (
	"encoding/json"
	"net/http"
)

type Train struct {
	Id      int64   `db:"id" json:"id"`
	Status  string  `db:"status" json:"status"`
	Acc     float64 `db:"acc" json:"acc"`
	Loss    float64 `db:"loss" json:"loss"`
	ValAcc  float64 `db:"val_acc" json:"val_acc"`
	ValLoss float64 `db:"val_loss" json:"val_loss"`
	Epochs  int     `db:"epochs" json:"epochs"`
	Name    string  `db:"name" json:"name"`
	Url     string  `db:"url" json:"url"` // saved model url
}

func (t *Train) Bind(r *http.Request) error {
	var body []byte
	_, err := r.Body.Read(body)
	if err != nil {
		return err
	}

	var train Train
	err = json.Unmarshal(body, &train)
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
	t.Epochs++
}

type TrainRepository interface {
	Insert(train Train) error
	Delete(opts ...Option) error
	Find(opts ...Option) (Train, error)
	FindAll(opts ...Option) ([]Train, error)
	Update(train Train, opts ...Option) error
}
