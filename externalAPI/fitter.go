package externalAPI

import (
	"encoding/json"
	"net/http"
)

type FitRequestBody struct {
	TrainId   int64                 `json:"train_id"`
	UserId    int64                 `json:"user_id"`
	ProjectNo int                   `json:"project_no"`
	Config    json.RawMessage       `json:"config"`
	Content   json.RawMessage       `json:"content"`
	DataSet   FitRequestBodyDataSet `json:"data_set"`
}

type FitRequestBodyDataSet struct {
	TrainUri      string                             `json:"train_uri"`
	ValidationUri string                             `json:"validation_uri"`
	Shuffle       bool                               `json:"shuffle"`
	Label         string                             `json:"label"`
	Normalization FitRequestBodyDataSetNormalization `json:"normalization"`
	Kind          string                             `json:"kind"`
}

type FitRequestBodyDataSetNormalization struct {
	Usage  bool   `json:"usage"`
	Method string `json:"method"`
}

type Fitter interface {
	Fit(payload FitRequestBody) (*http.Response, error)
}
