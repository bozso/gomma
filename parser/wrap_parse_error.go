package parser

import (
	"github.com/bozso/gomma/bit"
)

type Mode int

const (
	ModeInt Mode = iota
	ModeUint
	ModeFloat
)

func (m Mode) WrapError(ew ErrorWrapper, s string, e error) (err error) {
	if e != nil {
		e = ew.WrapParseError(s, m, e)
	}

	return e
}

type WrapError struct {
	p       Parser
	wrapper ErrorWrapper
}

func (we WrapError) ParseInt(s string, base bit.Base, size bit.Size) (ii int64, err error) {
	ii, err = we.p.ParseInt(s, base, size)
	err = ModeInt.WrapError(we.wrapper, s, err)

	return
}

func (we WrapError) ParseUint(s string, base bit.Base, size bit.Size) (ii uint64, err error) {
	ii, err = we.p.ParseUint(s, base, size)
	err = ModeUint.WrapError(we.wrapper, s, err)

	return
}

func (we WrapError) ParseFloat(s string, size bit.Size) (fl float64, err error) {
	fl, err = we.p.ParseFloat(s, size)
	err = ModeFloat.WrapError(we.wrapper, s, err)

	return
}
