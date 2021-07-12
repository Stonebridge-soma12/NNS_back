package service

import (
	"encoding/json"
	"net/http"
)

type responseBody map[string]interface{}

func writeJson(w http.ResponseWriter, code int, data interface{}) error {
	jsoned, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.WriteHeader(code)
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(jsoned)
	return err
}