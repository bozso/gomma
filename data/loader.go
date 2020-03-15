package data

import (
    "github.com/bozso/gomma/utils/params"    
    "github.com/bozso/gotoolbox/path"
)

type ParamKeys struct {
    RngKey, AziKey, TypeKey, DateKey string
}

var (
    DefaultKeys = ParamKeys{
        RngKey: "range_samples",
        AziKey: "azimuth_lines",
        TypeKey: "image_format",
        DateKey: "date",
    }
)

type PathWithPar struct {
    Path
    ParFile path.File
    keys *ParamKeys
}

    //if len(ext) == 0 {
        //ext = "par"
    //}
    
    //l.DatFile = file
    
    //// okay for parfile to not exist
    //l.ParFile = path.New(fmt.Sprintf("%s.%s", file, ext)).ToFile()


func (p Path) WithParFile(file path.File) (pp PathWithPar) {
    return PathWithPar{
        Path: p,
        ParFile: file,
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
    if err != nil { return }
    
    return pp.LoadWithParser(params.Parser{p})
}

func (pp PathWithPar) LoadWithParser(pr params.Parser) (f FileWithPar, err error) {
    k := pp.keys
    
    f.DatFile, err = pp.DatFile.ToValid()
    if err != nil {
        return
    }
    
    f.ParFile, err = pp.ParFile.ToValid()
    if err != nil {
        return
    }
    
    f.Ra.Rng, err = pr.Int(k.RngKey, 0)
    if err != nil { return }
    
    f.Ra.Azi, err = pr.Int(k.AziKey, 0)
    if err != nil { return }
    
    s, err := pr.Param(k.TypeKey)
    if err != nil { return }
    
    err = f.Dtype.Set(s)
    if err != nil { return }
    
    if d := k.DateKey; len(d) != 0 {
        s, err = pr.Param(d)
        if err != nil { return }
        
        f.Time, err = DateFmt.Parse(s)
    }
    
    return
}
