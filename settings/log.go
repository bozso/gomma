package settings

import (
	"io"
)

type Logger struct {
	writer io.WriteCloser
	level  int
}
