package dem

import (
    "github.com/bozso/gamma/data"
    "github.com/bozso/gamma/plot"
)

type File struct {
    data.File
}

func FromFile(path string) (d File, err error) {
    d.File, err = data.FromFile(path)
    return
}

func (d File) NewLookup(path string) (l Lookup) {
    l.Dat = path
    l.Ra = d.Ra
    l.dtype = FloatCpx
    return
}


func (d *File) Set(s string) (err error) {
    return
}

func (f File) ParseRng() (i int, err error) {
    i, err = f.Int("width", 0)
    return
}

func (f File) ParseAzi() (i int, err error) {
    i, err = f.Int("nlines", 0)
    return
}

func (f File) Raster(opt plot.RasArgs) (err error) {
    opt.Mode = plot.Power
    opt.Parse(f)
    
    err = plot.Raster(f, opt)
    return nil
}


