package externalAPI

import (
	"bytes"
	"encoding/json"
	"net/http"
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
	const requestUrl = "http://54.180.153.56:8080/fit"

	jsoned, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, requestUrl, bytes.NewBuffer(jsoned))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}
