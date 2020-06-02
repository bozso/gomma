package interferogram

import (
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/utils/params"
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
    p.PathWithPar = p.WithKeys(keys)
    
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
    par, err := p.ParFile.ToValid()
    if err != nil {
        return
    }

    diffPar, err := p.DiffPar.ToValid()
    if err != nil {
        return
    }
    
    ppar, err := data.NewGammaParams(par)
    if err != nil {
        return
    }

    pDiffPar, err := data.NewGammaParams(diffPar)
    if err != nil {
        return
    }

    parser := params.NewTeeParser(ppar, pDiffPar)

    
    fw, err := p.LoadWithParser(parser.ToParser())
    if err != nil {
        return
    }
    
    f.DatFile, f.ParFile = fw.DatFile, par
    f.DiffPar, f.Quality, f.SimUnwrap = diffPar, p.Quality, p.SimUnwrap
    
    err = f.Validate()
    
    return
}

// TODO: implement
var keys = &data.ParamKeys{
    Rng: "",
    Azi: "",
    Type: "",
    Date: "date",    
}
