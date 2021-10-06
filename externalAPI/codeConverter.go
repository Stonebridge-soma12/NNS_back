package externalAPI

import "net/http"

type CodeConvertRequestBody struct {
	Content interface{} `json:"content"`
	Config  interface{} `json:"config"`
}

type CodeConverter interface {
	CodeConvert(payload CodeConvertRequestBody) (*http.Response, error)
}
