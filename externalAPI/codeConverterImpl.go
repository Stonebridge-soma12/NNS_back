package externalAPI

import (
	"bytes"
	"encoding/json"
	"moul.io/http2curl"
	"net/http"
	"nns_back/log"
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
	const requetsUrl = "http://nnstudio.io:8081/api/python"

	jsoned, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, requetsUrl, bytes.NewBuffer(jsoned))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	command, err := http2curl.GetCurlCommand(req)
	if err != nil {
		return nil, err
	}

	log.Debug(command)

	return c.httpClient.Do(req)
}
