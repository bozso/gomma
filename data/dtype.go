package data

import (
    "fmt"
    "strings"
    
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

func (d *Type) SetCli(c *utils.Cli) {
    c.Var(d, "dtype", "Datatype of datafile.")
}

func (d *Type) Set(s string) error {
    in := strings.ToUpper(s)
    
    switch in {
    case "FLOAT":
        *d = Float
    case "DOUBLE":
        *d = Double
    case "SCOMPLEX":
        *d = ShortCpx
    case "FCOMPLEX":
        *d = FloatCpx
    case "SUN", "RASTER", "BMP":
        *d = Raster
    case "UNSIGNED CHAR":
        *d = UChar
    case "SHORT":
        *d = Short
    case "ANY":
        *d = Any
    default:
        *d = Unknown
    }
    
    return nil
}

func (d Type) String() string {
    switch d {
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
    default:
        return "UNKNOWN"
    }
}

type TypeMismatchError struct {
    datafile, expected string
    Type
    Err error
}

func (e TypeMismatchError) Error() string {
    return fmt.Sprintf("expected datatype(s) '%s' for datafile '%s', got '%s'",
        e.expected, e.datafile, e.Type.String())
}

func (e TypeMismatchError) Unwrap() error {
    return e.Err
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

func WrongType(dtype Type, kind string) error {
    return WrongTypeError{dtype, kind, nil}
}

type WrongTypeError struct {
    Type
    kind string
    err error
}

func (e WrongTypeError) Error() string {
    return fmt.Sprintf("wrong datatype '%s' for %s", e.Type.String(),
        e.kind)
}

func (e WrongTypeError) Unwrap() error {
    return e.err
}
