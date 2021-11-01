package externalAPI

import "net/http"

type CodeConvertRequestBody struct {
	TrainId int64       `json:"trainId"`
	Content interface{} `json:"content"`
	Config  interface{} `json:"config"`
}

type CodeConverter interface {
	CodeConvert(payload CodeConvertRequestBody) (*http.Response, error)
}
