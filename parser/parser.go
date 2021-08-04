package parser

import (
	"errors"
	"fmt"

	"github.com/bozso/gomma/bit"
)

type Parser interface {
	ParseInt(string, bit.Base, bit.Size) (int64, error)
	ParseUint(string, bit.Base, bit.Size) (uint64, error)
	ParseFloat(string, bit.Size) (float64, error)
	ParseBool(string) (bool, error)
}

/*
func (p Parser) Splitter(key string) (sp splitted.Parser, err error) {
	s, err := p.Param(key)
	if err != nil {
		return
	}

	sp, err = splitted.New(s, " ")
	return
}

func (p Parser) Int(key string, idx int) (ii int, err error) {
	sp, err := p.Splitter(key)
	if err != nil {
		return
	}

	ii, err = sp.Int(idx)
	return
}

func (p Parser) Float(key string, idx int) (ff float64, err error) {
	sp, err := p.Splitter(key)
	if err != nil {
		return
	}

	ff, err = sp.Float(idx)
	return
}
*/

/*
Wrapper for reading from many Retreivers. It will try to read
the appropriate parameters from each of the receivers. Returns
with error if parameter could not be found in any of them.
*/
type TeeGetter struct {
	getters []Getter
}

func NewTeeGetter(getters ...Getter) (t TeeGetter) {
	t.getters = getters
	return
}

var nf = NotFound{}

func (t TeeGetter) Get(key string) (s string, err error) {
	for _, r := range t.getters {
		s, err = r.Get(key)

		if err == nil {
			return
		}

		// try to retreive the parameter in the next retreiver
		if errors.Is(err, nf) {
			continue
		} else {
			return
		}
	}

	if len(s) == 0 {
		err = fmt.Errorf(
			"parameter '%s' was not found in any of the retreivers",
			key)
	}

	return
}
