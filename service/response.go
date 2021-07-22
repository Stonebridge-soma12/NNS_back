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
type additionalData interface {
	apply(*map[string]interface{})
}

type additionalDataFunc func(*map[string]interface{})

func (f additionalDataFunc) apply(m *map[string]interface{}) {
	f(m)
}

// col inserts additional data to the error message
func col(key string, value interface{}) additionalData {
	return additionalDataFunc(func(m *map[string]interface{}) {
		(*m)[key] = value
	})
}

const (
	target = "target"
)

func writeError(w http.ResponseWriter, code int, errMsg ErrMsg, additionalColumns ...additionalData) error {
	body := make(map[string]interface{})
	body["statusCode"] = code
	body["errMsg"] = errMsg

	for _, c := range additionalColumns {
		c.apply(&body)
	}

	return writeJson(w, code, body)
}
