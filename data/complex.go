package data

import (
    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gotoolbox/errors"

    "github.com/bozso/gomma/common"
)

type Complex struct {
    File
}

func (c Complex) Validate() (err error) {
    return c.EnsureComplex()
}

type CpxToReal int

const (
    Real CpxToReal = iota
    Imaginary
    Intensity
    Magnitude
    Phase
)

func (c CpxToReal) String() string {
    switch c {
    case Real:
        return "Real"
    case Imaginary:
        return "Imaginary"
    case Intensity:
        return "Intensity"
    case Magnitude:
        return "Magnitude"
    case Phase:
        return "Phase"
    default:
        return "Unknown"
    }
}

var cpxToReal = common.Must("cpx_to_real")

func (c Complex) ToReal(mode CpxToReal, file path.Path) (d File, err error) {
    Mode := 0
    
    switch mode {
    case Real:
        Mode = 0
    case Imaginary:
        Mode = 1
    case Intensity:
        Mode = 2
    case Magnitude:
        Mode = 3
    case Phase:
        Mode = 4
    default:
        err = errors.UnrecognizedMode(mode.String(), "Complex.ToReal")
        return
    }
    
    p := New(file)
    
    _, err = cpxToReal.Call(c.DatFile, p.DatFile, c.Ra.Rng, Mode)
    if err != nil {
        return
    }
    
    d, err = c.WithShapeDType(p, Float)
    return
}

type ComplexWithPar struct {
    Complex
    Parameter    
}

func (c *ComplexWithPar) Set(s string) (err error) {
    return LoadJson(s, c)
}
