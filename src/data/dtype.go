package data

import (
    "fmt"
    "strings"
    
    "../utils"
)

type DType int

const (
    Float DType = iota
    Double
    ShortCpx
    FloatCpx
    Raster
    UChar
    Short
    Unknown
    Any
)

func (d *DType) SetCli(c *utils.Cli) {
    c.Var(d, "dtype", "Datatype of datafile.")
}

func (d *DType) Set(s string) error {
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

func (d DType) String() string {
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
    ftype, expected string
    DType
    Err error
}

func (e TypeMismatchError) Error() string {
    return fmt.Sprintf("expected datatype '%s' for %s datafile, got '%s'",
        e.expected, e.ftype, e.DType.String())
}

func (e TypeMismatchError) Unwrap() error {
    return e.Err
}

type UnknownTypeError struct {
    DType
    Err error
}

func (e UnknownTypeError) Error() string {
    return fmt.Sprintf("unrecognised type '%s', expected a valid datatype",
        e.DType.String())
}

func (e UnknownTypeError) Unwrap() error {
    return e.Err
}

type WrongTypeError struct {
    DType
    kind string
    Err error
}

func (e WrongTypeError) Error() string {
    return fmt.Sprintf("wrong datatype '%s' for %s", e.kind, e.DType.String())
}

func (e WrongTypeError) Unwrap() error {
    return e.Err
}
