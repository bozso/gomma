package data

import (
	"fmt"
	"strings"
	"time"

	"github.com/bozso/gotoolbox/path"

	"github.com/bozso/gomma/common"
	"github.com/bozso/gomma/date"
)

const DateFmt date.ParseFmt = "2016 12 05"

type RngAzi struct {
	Rng uint
	Azi uint
}

type Meta struct {
	DataFile path.ValidFile `json:"datafile"`
	Dtype    Type           `json:"data_type"`
	RngAzi   RngAzi         `json:"rng_azi"`
	Time     time.Time      `json:"time"`
}

func (m *Meta) SetMeta(meta Meta) {
	*m = meta
}

func (m Meta) Date() time.Time {
	return m.Time
}

func (m Meta) DataType() Type {
	return m.Dtype
}

func (m Meta) TypeCheck(filepath path.Pather, dtypes ...Type) (err error) {
	D := m.Dtype

	for _, dt := range dtypes {
		if D == dt {
			return nil
		}
	}

	var sb strings.Builder

	for _, dt := range dtypes {
		sb.WriteString(dt.String() + ", ")
	}

	return TypeMismatchError{
		datafile: filepath,
		expected: sb.String(),
		Type:     D,
	}
}

func SameCols(one common.Dims, two common.Dims) *ShapeMismatchError {
	n1, n2 := one.Rng(), two.Rng()

	if n1 != n2 {
		return &ShapeMismatchError{
			n1:  n1,
			n2:  n2,
			dim: "range samples / columns",
		}
	}
	return nil
}

func SameRows(one, two RngAzi) (err *ShapeMismatchError) {
	n1, n2 := one.Azi, two.Azi

	if n1 != n2 {
		return &ShapeMismatchError{
			n1:  n1,
			n2:  n2,
			dim: "azimuth lines / rows",
		}
	}

	return
}

func SameShape(one, two RngAzi) (err *ShapeMismatchError) {
	err = SameCols(one, two)
	if err != nil {
		return
	}

	return SameRows(one, two)
}

type ShapeMismatchError struct {
	dat1, dat2 path.Pather
	dim        string
	n1, n2     uint
	err        error
}

func (s ShapeMismatchError) Error() string {
	return fmt.Sprintf(
		"expected datafile '%s' to have the same %s as datafile '%s' (%d != %d)", s.dat1, s.dim, s.dat2, s.n1, s.n2)
}

func (s ShapeMismatchError) WithPaths(one, two common.Pather) (sh ShapeMismatchError) {
	sh = s
	sh.dat1, sh.dat2 = one.Path(), two.Path()

	return s
}

func (s ShapeMismatchError) Unwrap() error {
	return s.err
}
