package geo

import (
    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/plot"
)


type Height struct {
    data.File
}

func (h Height) Validate() (err error) {
    return h.EnsureFloat()
}

func (h *Height) Set(s string) (err error) {
    return
}

func (h Height) Raster(opt plot.RasArgs) (err error) {
    opt.Mode = plot.Height
    opt.Parse(h)
    
    err = plot.Raster(h, opt)
    return
}
