package parser

import (
	"fmt"
)

type MissingKey struct {
	Key string
}

func (m MissingKey) Error() (s string) {
	return fmt.Sprintf("key '%s' not found", m.Key)
}
