package service

import (
	"encoding/json"
	"net/http"
)

// write response helper
type responseBody map[string]interface{}

func writeJson(w http.ResponseWriter, code int, data interface{}) error {
	jsoned, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(jsoned)
	return err
}

// write error response helper
type ErrBody struct {
	StatusCode int    `json:"statusCode"` // http status code
	ErrMsg     ErrMsg `json:"errMsg"`
}

func writeError(w http.ResponseWriter, code int, errMsg ErrMsg) error {
	body := ErrBody{
		StatusCode: code,
		ErrMsg:     errMsg,
	}
	return writeJson(w, code, body)
}
