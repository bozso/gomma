package data

import (
	"git.sr.ht/~istvan_bozso/sedet/bit"
	"git.sr.ht/~istvan_bozso/sedet/parser"
)

// Default metadate for parsing ints
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

	m.Date, err = DateFmt.Parse(s)

	return
}
