package externalAPI

import (
	"bytes"
	"encoding/json"
	"moul.io/http2curl"
	"net/http"
	"nns_back/log"
)

type fitterImpl struct {
	httpClient *http.Client
}

func NewFitter(httpClient *http.Client) Fitter {
	return &fitterImpl{
		httpClient: httpClient,
	}
}

func (c *fitterImpl) Fit(payload FitRequestBody) (*http.Response, error) {
	const requestUrl = "http://nnstudio.io:8081/api/fit"

	jsoned, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, requestUrl, bytes.NewBuffer(jsoned))
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
