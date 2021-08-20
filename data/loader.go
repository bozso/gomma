package data

import (
	"io"

	"git.sr.ht/~istvan_bozso/sedet/parser"
)

type GetterMaker interface {
	MakeGetter() parser.MutGetter
	PutGetter(parser.MutGetter)
}

type Loader struct {
	Maker  GetterMaker
	Parser parser.Parser
	Setup  parser.Setup
}

func (l Loader) LoadMeta(r io.Reader, pk ParamKeys) (m Meta, err error) {
	g := l.Maker.MakeGetter()
	defer l.Maker.PutGetter(g)

	err = l.Setup.ParseInto(r, g)
	if err != nil {
		return
	}

	m, err = pk.ParseMeta(g, l.Parser)

	return
}
