package util

import (
	"encoding/json"
	"net/http"
)

// write response helper
type ResponseBody map[string]interface{}

func WriteJson(w http.ResponseWriter, code int, data interface{}) error {
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
type AdditionalData interface {
	apply(*map[string]interface{})
}

type additionalDataFunc func(*map[string]interface{})

func (f additionalDataFunc) apply(m *map[string]interface{}) {
	f(m)
}

// KeyValue inserts additional data to the error message
func KeyValue(key string, value interface{}) AdditionalData {
	return additionalDataFunc(func(m *map[string]interface{}) {
		(*m)[key] = value
	})
}

const (
	target = "target"
)

func WriteError(w http.ResponseWriter, code int, errMsg ErrMsg, additionalColumns ...AdditionalData) error {
	body := make(map[string]interface{})
	body["statusCode"] = code
	body["errMsg"] = errMsg

	for _, c := range additionalColumns {
		c.apply(&body)
	}

	return WriteJson(w, code, body)
}
