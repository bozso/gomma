package sentinel1

import (
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/data"
    "github.com/bozso/gomma/utils/params"
)

const maxIW = 3

type(
    IWPath struct {
        data.PathWithPar
        TOPSpar path.File
    }
    
    IWPaths [maxIW]IWPath
)

func NewIW(data path.File) (p IWPath) {
    p.DatFile = data
    p.ParFile = data.AddExt("par")
    p.TOPSpar = data.AddExt("TOPS_par")
    
    return
}

func (p IWPath) WithTOPS(tops path.File) (pp IWPath) {
    p.TOPSpar = tops
    return p
}

func (p IWPath) GetParser() (pp params.Parser, err error) {
    p1, err := p.GetParser()
    if err != nil {
        return
    }
    
    p2, err := data.NewGammaParams(p.TOPSpar)
    if err != nil {
        return
    }
    
    pp = params.NewTeeParser(p1, p2).ToParser()
    return
}

func (p IWPath) Load() (iw IW, err error) {
    pp, err := p.GetParser()
    if err != nil {
        return
    }
    
    iw.FileWithPar, err = p.LoadWithParser(pp)
    iw.TOPSpar = p.TOPSpar
    return
}


type(  
    IW struct {
        data.ComplexWithPar
        TOPSPar path.ValidFile
    }

    IWs [maxIW]IW
)

func (iw IW) Move(dir path.Dir) (miw IW, err error) {
    miw = iw
    
    if miw.ComplexWithPar, err = iw.ComplexWithPar.Move(dir); err != nil {
        return
    }
    
    miw.TOPSpar, err = iw.TOPSpar.Move(dir)
    return
}
