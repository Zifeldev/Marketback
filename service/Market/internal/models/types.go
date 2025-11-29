package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type SizesJSON []string

func (s SizesJSON) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *SizesJSON) Scan(value interface{}) error {
	if value == nil {
		*s = SizesJSON{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return errors.New("failed to unmarshal JSONB value")
		}
		return json.Unmarshal(b, s)
	}
}
