package data

import (
	"git.sr.ht/~istvan_bozso/sedet/parser"
)

type MetaValidator interface {
	ValidateMeta(Meta) error
}

type ParseAndValidate struct {
	parser    MetaParser
	validator MetaValidator
}

func (pv ParseAndValidate) ParseMeta(g parser.Getter, p parser.Parser) (m Meta, err error) {
	m, err = pv.parser.ParseMeta(g, p)
	if err != nil {
		return
	}

	err = pv.validator.ValidateMeta(m)

	return
}
