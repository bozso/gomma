package data

import (
	"github.com/bozso/gomma/utils/params"
	"github.com/bozso/gotoolbox/path"
)

type ParamKeys struct {
	Rng, Azi, Type, Date string
}

var DefaultKeys = &ParamKeys{
	Rng:  "range_samples",
	Azi:  "azimuth_lines",
	Type: "image_format",
	Date: "date",
}

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

func (pp PathWithPar) WithKeys(keys *ParamKeys) PathWithPar {
	pp.keys = keys
	return pp
}

func (pp PathWithPar) WithParser(p params.Parser) (wp WithParser) {
	wp.PathWithPar, wp.parser = pp, p
	return
}

func (pp PathWithPar) GetParser() (p params.Params, err error) {
	par, err := pp.ParFile.ToValidFile()
	if err != nil {
		return
	}

	p, err = NewGammaParams(par)
	return
}

type Loadable interface {
	SetDataFile(path.ValidFile)
	SetParFile(path.ValidFile)
	SetMeta(Meta)
	Validate() error
}

func (pp PathWithPar) Load(l Loadable) (err error) {
	p, err := pp.GetParser()
	if err != nil {
		return
	}

	return pp.WithParser(p.ToParser()).Load(l)
}

type WithParser struct {
	PathWithPar
	parser params.Parser
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
