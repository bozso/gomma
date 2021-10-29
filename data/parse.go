package data

import (
	"github.com/bozso/gomma/date"

	"git.sr.ht/~istvan_bozso/sedet/bit"
	"git.sr.ht/~istvan_bozso/sedet/parser"
)

type MetaParser interface {
	ParseMeta(parser.Getter, parser.Parser) (Meta, error)
}

// Default metadata for parsing ints
var im = IntMeta{
	Base: bit.Base(10),
	Size: bit.IntSize(64),
}

/*ParamKeys holds the keys for different fields for Meta.*/
type ParamKeys struct {
	Range   string `json:"range"`
	Azimuth string `json:"azimuth"`
	Type    string `json:"type"`
	Date    string `json:"date"`
}

var DefaultKeys = ParamKeys{
	Range:   "range_samples",
	Azimuth: "azimuth_lines",
	Type:    "image_format",
	Date:    "date",
}

const DateParse = date.Format("2016 12 05")

func (pk ParamKeys) ParseMeta(g parser.Getter, p parser.Parser) (m Meta, err error) {
	pg := WithGetter(p, g)

	var (
		keys  = [2]string{pk.Range, pk.Azimuth}
		uints = [2]*uint64{&m.RngAzi.Rng, &m.RngAzi.Azi}
	)

	for ii := 0; ii < 2; ii++ {
		ui, err := pg.ParseUint(keys[ii], im.Base, im.Size)
		if err != nil {
			return m, err
		}

		*uints[ii] = ui
	}

	err = pg.ParseVar(pk.Type, &m.DataType)
	if err != nil {
		return
	}

	s, err := g.GetParsed(pk.Date)
	if err != nil {
		return
	}

	time, err := DateParse.ParseDate(s)
	if err != nil {
		return
	}
	m.Date = date.New(time)

	return
}
