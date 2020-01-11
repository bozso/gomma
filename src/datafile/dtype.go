package datafile

import (
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

