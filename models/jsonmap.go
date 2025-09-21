package models

import (
	"database/sql/driver"
	"encoding/json"
)

type JSONMap map[string]interface{}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte(`{}`), nil
	}
	return json.Marshal(m)
}
func (m *JSONMap) Scan(src interface{}) error {
	if src == nil {
		*m = JSONMap{}
		return nil
	}
	switch b := src.(type) {
	case []byte:
		return json.Unmarshal(b, m)
	case string:
		return json.Unmarshal([]byte(b), m)
	default:
		*m = JSONMap{}
		return nil
	}
}
