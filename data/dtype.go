package data

import (
	"fmt"
	"strings"

	"github.com/bozso/gotoolbox/cli"
)

type Type int

const (
	Float Type = iota
	Double
	ShortCpx
	FloatCpx
	Raster
	UChar
	Short
	Unknown
	Any
)

func (t *Type) SetCli(c *cli.Cli) {
	c.Var(t, "dtype", "Datatype of datafile.")
}

func (t *Type) Set(s string) error {
	in := strings.ToUpper(s)

	switch in {
	case "FLOAT":
		*t = Float
	case "DOUBLE":
		*t = Double
	case "SCOMPLEX":
		*t = ShortCpx
	case "FCOMPLEX":
		*t = FloatCpx
	case "SUN", "RASTER", "BMP":
		*t = Raster
	case "UNSIGNED CHAR":
		*t = UChar
	case "SHORT":
		*t = Short
	case "ANY":
		*t = Any
	default:
		*t = Unknown
	}

	return nil
}

func (t Type) String() string {
	switch t {
	case Float:
		return "FLOAT"
	case Double:
		return "DOUBLE"
	case ShortCpx:
		return "SCOMPLEX"
	case FloatCpx:
		return "FCOMPLEX"
	case Raster:
		return "RASTER"
	case UChar:
		return "UNSIGNED CHAR"
	case Short:
		return "SHORT"
	case Any:
		return "ANY"
	case Unknown:
		return "UNKNOWN"
	default:
		return "UNKNOWN"
	}
}

type TypeMismatchError struct {
	Expected string
	Got      Type
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
	Type
	Err error
}

func (e UnknownTypeError) Error() string {
	return fmt.Sprintf("unrecognised type '%s', expected a valid datatype",
		e.Type.String())
}

func (e UnknownTypeError) Unwrap() error {
	return e.Err
}

func (t Type) WrongType(purpose string) error {
	return WrongTypeError{t, purpose, nil}
}

func WrongType(dtype Type, kind string) error {
	return WrongTypeError{dtype, kind, nil}
}

type WrongTypeError struct {
	Type
	Kind string
	err  error
}

func (e WrongTypeError) Error() string {
	return fmt.Sprintf("wrong datatype '%s' for %s", e.Type.String(),
		e.Kind)
}

func (e WrongTypeError) Unwrap() error {
	return e.err
}
