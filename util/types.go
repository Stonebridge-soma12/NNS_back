package util

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
)

// Status shows the status of the row in the database.
type Status string

const (
	StatusNONE    Status = "" // No constraints
	StatusEXIST   Status = "EXIST"
	StatusDELETED Status = "DELETED"
)

// NullBytes represents a JSON that may be null.
// NullBytes implements the Scanner interface so
// it can be used as a scan destination
type NullBytes struct {
	Bytes []byte
	Valid bool
}

// Scan implements the Scanner interface.
func (nb *NullBytes) Scan(value interface{}) error {
	if value == nil {
		nb.Bytes, nb.Valid = nil, false
		return nil
	}
	nb.Valid = true

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprintf("failed to interface conversion"))
	}
	nb.Bytes = bytes
	return nil
}

// Value implements the driver Valuer interface.
func (nb NullBytes) Value() (driver.Value, error) {
	if !nb.Valid {
		return nil, nil
	}
	return nb.Bytes, nil
}

// NullJson represents a JSON that may be null.
// NullJson implements the Scanner interface so
// it can be used as a scan destination
type NullJson struct {
	Json  json.RawMessage
	Valid bool // Valid is true if Json is not NULL
}

// Scan implements the Scanner interface.
func (nj *NullJson) Scan(value interface{}) error {
	if value == nil {
		nj.Json, nj.Valid = nil, false
		return nil
	}
	nj.Valid = true

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprintf("failed to unmarshal Json value: %v", value))
	}

	result := json.RawMessage{}
	err := json.Unmarshal(bytes, &result)
	nj.Json = result
	return err
}

// Value implements the driver Valuer interface.
func (nj NullJson) Value() (driver.Value, error) {
	if !nj.Valid {
		return nil, nil
	}
	return nj.Json.MarshalJSON()
}
