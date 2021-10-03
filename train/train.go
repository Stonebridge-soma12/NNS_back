package train

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type Train struct {
	Id        int64   `db:"id" json:"id"`
	UserId    int64   `db:"user_id" json:"user_id"`
	TrainNo   int64     `db:"train_no" json:"train_no"`
	ProjectId int64   `db:"project_id" json:"project_id"`
	Status    string  `db:"status" json:"status"`
	Acc       float64 `db:"acc" json:"acc"`
	Loss      float64 `db:"loss" json:"loss"`
	ValAcc    float64 `db:"val_acc" json:"val_acc"`
	ValLoss   float64 `db:"val_loss" json:"val_loss"`
	Epochs    int     `db:"epochs" json:"epochs"`
	Name      string  `db:"name" json:"name"`
	ResultUrl string  `db:"result_url" json:"result_url"` // saved model url
}

type TrainConfig struct {
	Id                         int64           `db:"id" json:"id"`
	TrainId                    int64           `db:"train_id" json:"train_id"`
	TrainDatasetUrl            string          `db:"train_dataset_url" json:"train_dataset_url"`
	ValidDatasetUrl            sql.NullString  `db:"valid_dataset_url" json:"valid_dataset_url"`
	DatasetShuffle             bool            `db:"dataset_shuffle" json:"dataset_shuffle"`
	DatasetLabel               string          `db:"dataset_label" json:"dataset_label"`
	DatasetNormalizationUsage  bool            `db:"dataset_normalization_usage" json:"dataset_normalization_usage"`
	DatasetNormalizationMethod string          `db:"dataset_normalization_method" json:"dataset_normalization_method"`
	ModelContent               json.RawMessage `db:"model_content" json:"model_content"`
	ModelConfig                json.RawMessage `db:"model_config" json:"model_config"`
	CreateTime                 time.Time       `db:"create_time" json:"create_time"`
	UpdateTime                 time.Time       `db:"update_time" json:"update_time"`
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

