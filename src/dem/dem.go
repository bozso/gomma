package dem

import (
    "../data"
    "../plot"
)

type File struct {
    data.File
}

func New(dat, par string) (d DEM, err error) {
    var ferr = merr.Make("NewDEM")
    
    if d.DatParFile, err = NewDatParFile(dat, par, "par", Float);
       err != nil {
        err = ferr.Wrap(err)
    }
    
    return
}

func (d File) NewLookup(path string) (l Lookup) {
    l.Dat = path
    l.Ra = d.Ra
    l.DType = FloatCpx
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
    opt.Mode = Power
    opt.Parse(f)
    
    err = plot.Raster(f, opt)
    return nil
}


