package interferogram

import (
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/data"    
)

type Paths struct {
    data.PathWithPar
    DiffPar, Quality, SimUnwrap path.File
}

func New(dat path.Path) (p Paths) {
    p.DatFile = dat.ToFile()
    p.ParFile = dat.AddExt("off").ToFile()
    p.DiffPar = dat.AddExt("diff_par").ToFile()
    p.Quality = dat.AddExt("qual").ToFile()
    p.SimUnwrap = dat.AddExt("sim_unwrap").ToFile()
    
    return
}

func (p Paths) WithParFile(file path.Path) (pp Paths) {
    p.ParFile = file.ToFile()
    return p
}

func (p Paths) WithDiffPar(file path.Path) (pp Paths) {
    p.DiffPar = file.ToFile()
    return p
}

func (p Paths) WithQuality(file path.Path) (pp Paths) {
    p.Quality = file.ToFile()
    return p
}

func (p Paths) WithSimUnwrap(file path.Path) (pp Paths) {
    p.SimUnwrap = file.ToFile()
    return p
}

// TODO: implement
func (p Paths) Load() (f File, err error) {
    fw, err := p.PathWithPar.Load()
    f.File, f.Parameter = fw.File, fw.Parameter
    
    return
}

// TODO: implement
var Importer = data.ParamKeys{
    
}
