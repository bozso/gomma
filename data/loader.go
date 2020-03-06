package data

import (
    "fmt"
    
    "github.com/bozso/gamma/utils/params"    
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

type Loader struct {
    DatFile, ParFile string
    keys *ParamKeys
    p params.Parser
}

func FromDataPath(p, ext string) (l Loader) {
    if len(ext) == 0 {
        ext = "par"
    }
    
    l.DatFile = p
    l.ParFile = fmt.Sprintf("%s.%s", p, ext)
    l.keys = &DefaultKeys
    return
}

func (l Loader) WithKeys(keys *ParamKeys) Loader {
    l.keys = keys
    return l
}

func (l Loader) GetParser() (p params.Params, err error) {
    p, err = NewGammaParams(l.ParFile)
    return
}

func (l Loader) Load() (f File, err error) {
    p, err := l.GetParser()
    if err != nil { return }
    
    return l.LoadWithParser(params.Parser{p})
}

func (l Loader) LoadWithParser(pr params.Parser) (f File, err error) {
    k := l.keys
    
    f.DatFile, f.ParFile = l.DatFile, l.ParFile
    
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
