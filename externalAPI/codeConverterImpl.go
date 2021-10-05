package externalAPI

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type codeConverterImpl struct {
	httpClient *http.Client
}

func NewCodeConverter(httpClient *http.Client) CodeConverter {
	return &codeConverterImpl{
		httpClient: httpClient,
	}
}

func (c *codeConverterImpl) CodeConvert(payload CodeConvertRequestBody) (*http.Response, error) {
	const requetsUrl = "http://54.180.153.56:8080/make-python"

	jsoned, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, requetsUrl, bytes.NewBuffer(jsoned))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}
