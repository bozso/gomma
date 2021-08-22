package data

import (
	"io"
	"io/fs"

	"git.sr.ht/~istvan_bozso/sedet/parser"
)

type GetterMaker interface {
	MakeGetter() parser.MutGetter
	PutGetter(parser.MutGetter)
}

type Loader struct {
	fsys   fs.FS
	maker  GetterMaker
	parser parser.Parser
	setup  parser.Setup
}

func (l Loader) ParseMeta(r io.Reader, mp MetaParser) (m Meta, err error) {
	g := l.maker.MakeGetter()
	defer l.maker.PutGetter(g)

	err = l.setup.ParseInto(r, g)
	if err != nil {
		return
	}

	m, err = mp.ParseMeta(g, l.parser)

	return
}

func (l Loader) MetaFromFile(path string, pk ParamKeys) (m Meta, err error) {
	f, err := l.fsys.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	m, err = l.ParseMeta(f, pk)

	return
}

func (l Loader) LoadFile(p PathWithPar, pk ParamKeys) (f File, err error) {
	meta, err := l.MetaFromFile(p.ParFile, pk)
	if err != nil {
		return
	}

	return File{
		Meta:     meta,
		DataFile: p.Path,
	}, nil
}
