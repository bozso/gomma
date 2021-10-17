package data

import (
	"io"
	"io/fs"
	"sync"

	"git.sr.ht/~istvan_bozso/sedet/parser"
	sfs "git.sr.ht/~istvan_bozso/shutil/fs"
)

type GetterMaker interface {
	MakeGetter() parser.MutGetter
	PutGetter(parser.MutGetter)
}

type GetterPool struct {
	pool sync.Pool
}

func NewGetterPool() (gp GetterPool) {
	return GetterPool{
		pool: sync.Pool{
			New: func() (v interface{}) {
				return parser.NewMap()
			},
		},
	}
}

func (m *GetterPool) MakeGetter() (m parser.MutGetter) {
	return pool.Get().(parser.MutGetter)
}

func (m *GetterPool) PutGetter(m parser.MutGetter) {
	return pool.Put(m)
}

func DefaultLoader() (l Loader) {
	return Loader{
		fsys:  sfs.OS(),
		maker: NewGetterPool(),
	}
}

type Loader struct {
	fsys   fs.FS
	maker  GetterMaker
	parser parser.Parser
	setup  parser.Setup
}

type UseParamFunc func(parser.Getter) error

func (l Loader) OpenParamGetter(path string, fn UseParamFunc) (err error) {
	f, err := l.fsys.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	return l.WithParamGetter(f, fn)
}

func (l Loader) WithParamGetter(r io.Reader, fn UseParamFunc) (err error) {
	g := l.maker.MakeGetter()
	defer l.maker.PutGetter(g)

	err = l.setup.ParseInto(r, g)
	if err != nil {
		return
	}

	return fn(g)
}

func (l Loader) ParseMeta(r io.Reader, mp MetaParser) (m Meta, err error) {
	l.WithParamGetter(r, func(g parser.Getter) (err error) {
		m, err = mp.ParseMeta(g, l.parser)
		return
	})

	return
}

func (l Loader) MetaFromFile(path string, mp MetaParser) (m Meta, err error) {
	f, err := l.fsys.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	m, err = l.ParseMeta(f, mp)

	return
}

func (l Loader) LoadFile(p PathWithPar, mp MetaParser) (f File, err error) {
	meta, err := l.MetaFromFile(p.ParFile, mp)
	if err != nil {
		return
	}

	return File{
		Meta:     meta,
		DataFile: p.Path,
	}, nil
}
