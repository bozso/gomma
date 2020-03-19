package base

import (
    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/plot"
)

type MLI struct {
    data.FileWithPar `json:"MLI"`
}

func (m MLI) Validate() (err error) {
    return m.EnsureFloat()
}

func (m MLI) Raster(opt plot.RasArgs) (err error) {
    opt.Mode = plot.Power
    return m.Raster(opt)
}
