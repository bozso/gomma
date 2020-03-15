package dem

import (
    "fmt"

    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/plot"

    "github.com/bozso/gotoolbox/path"
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

type PathWithPar struct {
    data.PathWithPar
}

func NewWithPar(dat, par path.File) (p PathWithPar) {
    p.PathWithPar = data.New(dat).WithParFile(par).WithKeys(&Import)
    return
}

func New(file path.File) (p PathWithPar) {
    par := path.New(fmt.Sprintf("%s.dem_par", file)).ToFile()
    
    return NewWithPar(file, par)
}

func (p PathWithPar) Load() (f File, err error) {
    f, err = p.Load()
    return
}
func (f File) Raster(opt plot.RasArgs) (err error) {
    opt.Mode = plot.Power
    opt.Parse(f)
    
    err = plot.Raster(f, opt)
    return nil
}
