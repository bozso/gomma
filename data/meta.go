package data

import (
	"fmt"
	"strings"
	"time"

	"github.com/bozso/gomma/date"
)

const DateFmt date.ParseFmt = "2016 12 05"

type Meta struct {
	DataType Type      `json:"datatype"`
	RngAzi   RngAzi    `json:"range_azimuth"`
	Date     time.Time `json:"date"`
}

func (m Meta) IsType(dtypes ...Type) (b bool) {
	D := m.Dtype

	for _, dt := range dtypes {
		if D == dt {
			return true
		}
	}
	return false
}

func (m Meta) MustBeOfType(dtypes ...Type) (err error) {
	if m.IsType(dtypes...) {
		return nil
	}

	var sb strings.Builder

	for _, dt := range dtypes {
		fmt.Fprintf(sb, "%s, ", dt.String())
	}

	return TypeMismatchError{
		Expected: sb.String(),
		Got:      m.Dtype,
	}
}
