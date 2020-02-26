package dem

import (
    "github.com/bozso/gamma/data"
    "github.com/bozso/gamma/plot"
)

type File struct {
    data.File
}

var Import = data.Importer{
    RngKey: "width",
    AziKey: "nlines",
}

func FromFile(path string) (d File, err error) {
    err = d.Set(path)
    if err != nil { return; }
    
    err = d.TypeCheck("DEM", "float", data.Float)
    return

}

func (d File) NewLookup(path string) (l Lookup) {
    l.DatFile = path
    l.Ra = d.Ra
    l.Dtype = data.FloatCpx
    return
}


func (d *File) Set(s string) (err error) {
    return
}

func (f File) Raster(opt plot.RasArgs) (err error) {
    opt.Mode = plot.Power
    opt.Parse(f)
    
    err = plot.Raster(f, opt)
    return nil
}
