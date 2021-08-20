package data

import (
	"git.sr.ht/~istvan_bozso/sedet/bit"
	"git.sr.ht/~istvan_bozso/sedet/parser"
)

type IntMeta struct {
	Size bit.Size
	Base bit.Base
}

func (im IntMeta) UintVar(p parser.Parser) (pu parsableUint) {
	return parsebleUint{
		value:  0,
		parser: p,
		Meta:   im,
	}
}

func (im IntMeta) IntVar(p parser.Parser) (pi parsableUint) {
	return parsebleInt{
		value:  0,
		parser: p,
		Meta:   im,
	}
}

type parsableUint struct {
	value  uint64
	parser parser.Parser
	Meta   IntMeta
}

func (pu *parsableUint) Set(s string) (err error) {
	ui, err := pu.parser.ParseUint(s, pu.Meta.Base, pu.Meta.Size)
	if err != nil {
		return
	}

	pu.value = ui
}

type parsableInt struct {
	value  int64
	parser parser.Parser
	Meta   IntMeta
}

func (pi *parsableInt) Set(s string) (err error) {
	ii, err := pu.parser.ParseInt(s, pi.Meta.Base, pi.Meta.Size)
	if err != nil {
		return
	}

	pi.value = ii
}

type parsableFloat struct {
	value  float64
	parser parser.Parser
	Size   bit.Size
}

func (pf *parsableFloat) Set(s string) (err error) {
	fl, err := pu.parser.ParseFloat(s, pf.Meta.Size)
	if err != nil {
		return
	}

	pf.value = fl
}
