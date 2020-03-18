package data

import (
    "github.com/bozso/gomma/common"
    "github.com/bozso/gomma/utils/params"
    "github.com/bozso/gotoolbox/path"
)

type ParamKeys struct {
    Rng, Azi, Type, Date string
}

var (
    DefaultKeys = ParamKeys{
        Rng: "range_samples",
        Azi: "azimuth_lines",
        Type: "image_format",
        Date: "date",
    }
)

type PathWithPar struct {
    Path
    ParFile path.File
    keys *ParamKeys
}

func (p Path) WithParFile(file path.Path) (pp PathWithPar) {
    return PathWithPar{
        Path: p,
        ParFile: file.ToFile(),
        keys: &DefaultKeys,
    }
}

func (pp PathWithPar) WithKeys(keys *ParamKeys) PathWithPar {
    pp.keys = keys
    return pp
}

func (pp PathWithPar) GetParser() (p params.Params, err error) {
    par, err := pp.ParFile.ToValid()
    if err != nil {
        return
    }
    
    p, err = NewGammaParams(par)
    return
}

func (pp PathWithPar) Load() (f FileWithPar, err error) {
    p, err := pp.GetParser()
    if err != nil {
        return
    }
    
    return pp.LoadWithParser(p.ToParser())
}

func (pp PathWithPar) LoadWithParser(pr params.Parser) (f FileWithPar, err error) {
    k := pp.keys
    
    f.ParFile, err = pp.ParFile.ToValid()
    if err != nil {
        return
    }
    
    ra := common.RngAzi{}
    
    ra.Rng, err = pr.Int(k.Rng, 0)
    if err != nil {
        return
    }
    
    ra.Azi, err = pr.Int(k.Azi, 0)
    if err != nil {
        return
    }
    
    s, err := pr.Param(k.Type)
    if err != nil {
        return
    }
    
    var dt Type
    err = dt.Set(s)
    if err != nil {
        return
    }
    
    if d := k.Date; len(d) != 0 {
        s, err = pr.Param(d)
        if err != nil {
            return
        }
        
        f.Time, err = DateFmt.Parse(s)
    }

    f.File, err = pp.Path.Load(ra, dt)
    return
}
