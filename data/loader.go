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

/*

import (
	"github.com/bozso/gomma/utils/params"
	"github.com/bozso/gotoolbox/path"
)

type PathWithPar struct {
	Path
	ParFile path.Path
	keys    *ParamKeys
}

func (p PathWithPar) WithPar(file path.Path) (pp PathWithPar) {
	p.ParFile = file
	return p
}

func (p Path) WithParFile(file path.Path) (pp PathWithPar) {
	return PathWithPar{
		Path:    p,
		ParFile: file,
		keys:    DefaultKeys,
	}
}

func (pp PathWithPar) WithParser(p params.Parser) (wp WithParser) {
	wp.PathWithPar, wp.parser = pp, p
	return
}

func (pp PathWithPar) Load(l Loadable) (err error) {
	p, err := pp.GetParser()
	if err != nil {
		return
	}

	return pp.WithParser(p.ToParser()).Load(l)
}

func (pp WithParser) Load(l Loadable) (err error) {
	f, err := pp.DataFile.ToValidFile()
	if err != nil {
		return
	}
	l.SetDataFile(f)

	f, err = pp.ParFile.ToValidFile()
	if err != nil {
		return
	}
	l.SetParFile(f)

	pr, k := pp.parser, pp.keys

	meta := Meta{}
	meta.RngAzi.Rng, err = pr.Int(k.Rng, 0)
	if err != nil {
		return
	}

	meta.RngAzi.Azi, err = pr.Int(k.Azi, 0)
	if err != nil {
		return
	}

	s, err := pr.Param(k.Type)
	if err != nil {
		return
	}

	err = meta.Dtype.Set(s)
	if err != nil {
		return
	}

	if d := k.Date; len(d) != 0 {
		s, err = pr.Param(d)
		if err != nil {
			return
		}

		meta.Time, err = DateFmt.Parse(s)
	}

	l.SetMeta(meta)

	err = l.Validate()
	return
}
*/
