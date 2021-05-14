package service

import (
	"github.com/bozso/gomma/dem"
	"github.com/bozso/gomma/geo"
	"github.com/bozso/gomma/mli"
	"github.com/bozso/gomma/plot"
	"github.com/bozso/gomma/slc"
)

type Plottable struct {
	plot.Plottable
}

var plottables = [...]plot.Plottable{
	mli.MLI{},
	mli.SLC{},
	dem.File{},
	geo.Height{},
}

func (p *Plottable) UnmarshalJSON(b []byte) (err error) {
	for _, candiate := range plottables {
		p.Plottable = candidate
		if err = json.Unmarshal(b, &p.Plottable); err == nil {
			return
		}
	}

	return
}
