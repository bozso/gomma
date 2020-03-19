package geo

import (
    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/plot"
)


type Hgt struct {
    data.File
}

func (h Hgt) Validate() (err error) {
    return h.EnsureFloat()
}

func (h *Hgt) Set(s string) (err error) {
    return
}

func (h Hgt) Raster(opt plot.RasArgs) (err error) {
    opt.Mode = plot.Height
    opt.Parse(h)
    
    err = plot.Raster(h, opt)
    return nil
}
