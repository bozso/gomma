package dem

import (
    "github.com/bozso/gamma/data"
    "github.com/bozso/gamma/plot"
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

func FromDataPath(p string) (f File) {
    f.FloatFile.File = data.FromDataPath(p, "dem_par")
    return
}

func (f *File) Load() (err error) {
    return f.FloatFile.Load(&Import)
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
