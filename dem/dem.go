package dem

import (
    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/plot"
)

type File struct {
    data.FloatFile
}

var Import = data.ParamKeys{
    RngKey: "width",
    AziKey: "nlines",
    TypeKey: "",
    DateKey: "",
}

type Loader struct {
    data.Loader
}

func FromDataPath(p string) (l Loader) {
    l.Loader = data.FromDataPath(p, "dem_par").WithKeys(&Import)
    return
}

func (l Loader) Load() (f File, err error) {
    f, err = l.Load()
    return
}

func (d File) NewLookup(path string) (l Lookup) {
    l.DatFile = path
    l.Ra = d.Ra
    l.Dtype = data.FloatCpx
    return
}

func (f File) Raster(opt plot.RasArgs) (err error) {
    opt.Mode = plot.Power
    opt.Parse(f)
    
    err = plot.Raster(f, opt)
    return nil
}
