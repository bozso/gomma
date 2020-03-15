package base

import (
    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/plot"
)

type MLI struct {
    data.FloatFile `json:"MLI"`
}

func (m MLI) Raster(opt plot.RasArgs) (err error) {
    opt.Mode = plot.Power
    return m.Raster(opt)
}
