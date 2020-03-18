package dem

import (
    "fmt"

    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/plot"

    "github.com/bozso/gotoolbox/path"
)

type File struct {
    data.FloatFileWithPar
}

var Keys = data.ParamKeys{
    Rng: "width",
    Azi: "nlines",
    Type: "",
    Date: "",
}

type PathWithPar struct {
    data.PathWithPar
}

func NewWithPar(dat, par path.Path) (p PathWithPar) {
    p.PathWithPar = data.New(dat).WithParFile(par).WithKeys(&Keys)
    return
}

func New(file path.Path) (p PathWithPar) {
    par := path.New(fmt.Sprintf("%s.dem_par", file))    
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
