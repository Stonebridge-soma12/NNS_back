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
type ErrCode string

const (
	// 400
	ErrDuplicate          ErrCode = "Duplicate Entity"
	ErrInvalidPathParm    ErrCode = "Invalid Path Parameter"
	ErrInvalidQueryParm   ErrCode = "Invalid Query Parameter"
	ErrInvalidRequestBody ErrCode = "Invalid Request Body"
	ErrBadRequest         ErrCode = "Bad Request"
	ErrNotFound           ErrCode = "Not Found"

	// 500
	ErrInternalServerError ErrCode = "Internal Server Error"
)

type ErrBody struct {
	StatusCode int     `json:"statusCode"` // http status code
	ErrMsg     ErrCode `json:"errMsg"`
}

func writeError(w http.ResponseWriter, code int, errCode ErrCode) error {
	body := ErrBody{
		StatusCode: code,
		ErrMsg:     errCode,
	}
	return writeJson(w, code, body)
}
