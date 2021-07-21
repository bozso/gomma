package data

import (
	"strings"

	"github.com/bozso/gotoolbox/enum"
)

// Complex represents a datafile holding complex values.
type Complex struct {
	File
}

// AsFile implements the interface.
func (c Complex) AsFile() (f File) {
	return c.File
}

// Validate ensures the datatype is complex.
func (c Complex) Validate() (err error) {
	return c.File.Meta.MustBeComplex()
}

// CpxToReal describes the conversion mode.
type CpxToReal int

const (
	ToReal CpxToReal = iota
	ToImaginary
	ToIntensity
	ToMagnitude
	ToPhase
)

var cpxToReal = enum.NewStringSet("ToReal", "ToImaginary",
	"ToIntensity", "ToMagnitude", "ToPhase").EnumType("CpxToReal")

func (c *CpxToReal) Set(str string) (err error) {
	switch strings.ToLower(str) {
	case "toreal":
		*c = ToReal
	case "toimaginary":
		*c = ToImaginary
	case "tointensity":
		*c = ToIntensity
	case "tomagnitude":
		*c = ToMagnitude
	case "tophase":
		*c = ToPhase
	default:
		err = cpxToReal.UnknownElement(str)
	}
	return
}

func (c CpxToReal) String() (s string) {
	switch c {
	case ToReal:
		s = "ToReal"
	case ToImaginary:
		s = "ToImaginary"
	case ToIntensity:
		s = "ToIntensity"
	case ToMagnitude:
		s = "ToMagnitude"
	case ToPhase:
		s = "ToPhase"
	default:
		s = "Unknown"
	}
	return
}

/*
func (c Complex) ComplexToReal(cmd command.Command, mode CpxToReal, dst path.Path) (r Real, err error) {
	Mode := 0

	switch mode {
	case ToReal:
		Mode = 0
	case ToImaginary:
		Mode = 1
	case ToIntensity:
		Mode = 2
	case ToMagnitude:
		Mode = 3
	case ToPhase:
		Mode = 4
	default:
		err = cpxToReal.UnknownElement(mode.String())
		return
	}

	p := New(dst)

	_, err = cmd.Call(c.DataFile, p.DataFile, c.Rng, Mode)
	if err != nil {
		return
	}

	r.File, err = c.WithShapeDType(p, Float)
	return
}
*/

type ComplexWithPar struct {
	Complex
	Parameter `json:"parameter"`
}
