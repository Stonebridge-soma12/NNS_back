package service

import (
	"encoding/json"
	"io"
)

type validator interface {
	validate() error
}

func bindJson(r io.Reader, v validator) error {
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		return err
	}

	return v.validate()
}