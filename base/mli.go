package base

import (
    "github.com/bozso/gamma/data"
    "github.com/bozso/gamma/plot"
)

type MLI struct {
    data.FloatFile `json:"MLI"`
}

func (m MLI) Raster(opt plot.RasArgs) (err error) {
    opt.Mode = plot.Power
    return m.Raster(opt)
}
