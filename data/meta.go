package data

import (
	"fmt"
	"strings"
	"time"

	"github.com/bozso/gomma/date"
)

const DateFmt date.ParseFmt = "2016 12 05"

type Meta struct {
	DataType  Type      `json:"data_type"`
	RngAzi    RngAzi    `json:"range_azimuth"`
	Date      time.Time `json:"date"`
	CreatedBy CreatedBy `json:"created_by"`
}

func (m Meta) IsComplex() (b bool) {
	return m.IsType(FloatCpx, ShortCpx)
}

func (m Meta) IsReal() (b bool) {
	return m.IsType(Float, Double)
}

func (m Meta) MustBeComplex() (err error) {
	return m.MustBeOfType(FloatCpx, ShortCpx)
}

func (m Meta) MustBeReal() (err error) {
	return m.MustBeOfType(Float, Double)
}

func (m Meta) IsType(dtypes ...Type) (b bool) {
	D := m.DataType

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

	sb := &strings.Builder{}

	for _, dt := range dtypes {
		fmt.Fprintf(sb, "%s, ", dt.String())
	}

	return TypeMismatchError{
		Expected: sb.String(),
		Got:      m.DataType,
	}
}
