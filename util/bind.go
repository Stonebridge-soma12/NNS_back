package util

import (
	"encoding/json"
	"io"
)

type Validator interface {
	Validate() error
}

func BindJson(r io.Reader, v Validator) error {
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		return err
	}

	return v.Validate()
}