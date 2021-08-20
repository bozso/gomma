package data

import (
	"git.sr.ht/~istvan_bozso/sedet/bit"
	"git.sr.ht/~istvan_bozso/sedet/parser"
)

type Var interface {
	Set(string) error
}

type ParserWithGetter struct {
	parser parser.Parser
	getter parser.Getter
}

func WithGetter(p parser.Parser, g parser.Getter) (pg ParserWithGetter) {
	return ParserWithGetter{
		parser: p,
		getter: g,
	}
}

func (p ParserWithGetter) ParseVar(key string, v Var) (err error) {
	s, err := p.getter.GetParsed(key)
	if err != nil {
		return
	}

	err = v.Set(s)

	return
}

func (p ParserWithGetter) ParseUint(key string, base bit.Base, size bit.Size) (ui uint64, err error) {
	v := &IntMeta{base, size}.UintVar(p.parser)
	err = p.ParseVar(key, &v)
	if err != nil {
		return
	}

	ui = v.value

	return
}

func (p ParserWithGetter) ParseInt(key string, base bit.Base, size bit.Size) (ii int64, err error) {
	v := &IntMeta{base, size}.IntVar(p.parser)
	err = p.ParseVar(key, &v)
	if err != nil {
		return
	}

	ii = v.value

	return
}

func (p ParserWithGetter) ParseFloat(key string, size bit.Size) (fl float64, err error) {
	v := ParseFloat{parser: p.parser, Size: size}
	err = p.ParseVar(key, &v)
	if err != nil {
		return
	}

	fl = v.value

	return
}
