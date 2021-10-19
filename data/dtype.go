package data

import (
	"fmt"
	"strings"

	"github.com/bozso/gotoolbox/cli"
)

type Kind int

const (
	KindFloat Kind = iota
	KindDouble
	KindShortCpx
	KindFloatCpx
	KindRaster
	KindUChar
	KindShort
	KindUnknown
	KindAny
)

func (k *Kind) SetCli(c *cli.Cli) {
	c.Var(k, "dtype", "Datatype of datafile.")
}

func (t *Kind) Set(s string) error {
	in := strings.ToUpper(s)

	switch in {
	case "FLOAT":
		*t = KindFloat
	case "DOUBLE":
		*t = KindDouble
	case "SCOMPLEX":
		*t = KindShortCpx
	case "FCOMPLEX":
		*t = KindFloatCpx
	case "SUN", "RASTER", "BMP":
		*t = KindRaster
	case "UNSIGNED CHAR":
		*t = KindUChar
	case "SHORT":
		*t = KindShort
	case "ANY":
		*t = KindAny
	default:
		*t = KindUnknown
	}

	return nil
}

func (k Kind) String() string {
	switch k {
	case KindFloat:
		return "FLOAT"
	case KindDouble:
		return "DOUBLE"
	case KindShortCpx:
		return "SCOMPLEX"
	case KindFloatCpx:
		return "FCOMPLEX"
	case KindRaster:
		return "RASTER"
	case KindUChar:
		return "UNSIGNED CHAR"
	case KindShort:
		return "SHORT"
	case KindAny:
		return "ANY"
	case KindUnknown:
		return "UNKNOWN"
	default:
		return "UNKNOWN"
	}
}

type TypeMismatchError struct {
	Expected string
	Got      Kind
	err      error
}

func (e TypeMismatchError) Error() string {
	return fmt.Sprintf("expected datatype(s) '%s', got '%s'",
		e.Expected, e.Got)
}

func (e TypeMismatchError) Unwrap() error {
	return e.err
}

type UnknownTypeError struct {
	Kind
	Err error
}

func (e UnknownTypeError) Error() string {
	return fmt.Sprintf("unrecognised type '%s', expected a valid datatype",
		e.Kind.String())
}

func (e UnknownTypeError) Unwrap() error {
	return e.Err
}

func (k Kind) WrongType(purpose string) error {
	return WrongTypeError{Kind: k, data: purpose, err: nil}
}

func WrongType(dtype Kind, data string) error {
	return WrongTypeError{Kind: dtype, data: data, err: nil}
}

type WrongTypeError struct {
	Kind
	data string
	err  error
}

func (e WrongTypeError) Error() string {
	return fmt.Sprintf("wrong datatype '%s' for %s", e.Kind.String(),
		e.data)
}

func (e WrongTypeError) Unwrap() error {
	return e.err
}
