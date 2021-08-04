package parser

import (
	"strconv"

	"github.com/bozso/gomma/bit"
)

type Default struct{}

func (Default) ParseInt(s string, base bit.Base, size bit.Size) (ii int64, err error) {
	ii, err = strconv.ParseInt(s, int(base), int(size))

	return
}

func (Default) ParseUInt(s string, base bit.Base, size bit.Size) (ii uint64, err error) {
	ii, err = strconv.ParseUint(s, int(base), int(size))

	return
}

func (Default) ParseFloat(s string, size bit.Size) (ff float64, err error) {
	ff, err = strconv.ParseFloat(s, int(size))

	return
}
