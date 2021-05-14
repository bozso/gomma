package meta

import (
	"encoding/json"
)

type Config struct {
	Tag  string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}
