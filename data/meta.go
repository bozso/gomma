package data

import (
	"fmt"
	"strings"

	"github.com/bozso/gomma/date"
)

const DateFmt date.ParseFmt = "2016 12 05"

type Meta struct {
	DataType  Kind      `json:"data_type"`
	RngAzi    RngAzi    `json:"range_azimuth"`
	Date      date.Date `json:"date"`
	CreatedBy CreatedBy `json:"created_by"`
}

func (m Meta) IsComplex() (b bool) {
	return m.IsType(KindFloatCpx, KindShortCpx)
}

func (m Meta) IsReal() (b bool) {
	return m.IsType(KindFloat, KindDouble)
}

func (m Meta) MustBeComplex() (err error) {
	return m.MustBeOfType(KindFloatCpx, KindShortCpx)
}

func (m Meta) MustBeReal() (err error) {
	return m.MustBeOfType(KindFloat, KindDouble)
}

func (m Meta) IsType(dtypes ...Kind) (b bool) {
	D := m.DataType

	for _, dt := range dtypes {
		if D == dt {
			return true
		}
	}
	return false
}

func (m Meta) MustBeOfType(dtypes ...Kind) (err error) {
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
