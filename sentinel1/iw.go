package sentinel1

import (
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/utils/params"
)

const maxIW = 3

type(
    IWLoader struct {
        data.Loader
        TOPS_par string
    }
    
    IWLoaders [maxIW]IWLoader
)

func NewIW(dat, par, TOPS_par string) (l IWLoader) {
    l.DatFile = dat
    
    if len(par) == 0 {
        par = dat + ".par"
    }
    
    l.ParFile = par
    
    if len(TOPS_par) == 0 {
        TOPS_par = dat + ".TOPS_par"
    }

    l.TOPS_par = TOPS_par
    
    // do we need a custom importer?
    
    return
}

func (l IWLoader) GetParser() (p params.Parser, err error) {
    p1, err := l.GetParser()
    if err != nil { return }
    
    p2, err := data.NewGammaParams(l.TOPS_par)
    if err != nil { return }
    
    p = params.NewTeeParser(p1, p2).ToParser()
    return
}

func (l IWLoader) Load() (iw IW, err error) {
    p, err := l.GetParser()
    if err != nil { return }
    
    iw.File, err = l.LoadWithParser(p)
    iw.TOPS_par = l.TOPS_par
    return
}


type(  
    IW struct {
        data.ComplexFile
        TOPS_par string
    }

    IWs [maxIW]IW
)

func (iw IW) Move(dir string) (miw IW, err error) {
    if miw.File, err = iw.File.Move(dir); err != nil {
        return
    }
    
    miw.TOPS_par, err = path.Move(iw.TOPS_par, dir)
    return
}
