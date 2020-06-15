package sentinel1

import (
    "fmt"
    
    "github.com/bozso/gotoolbox/path"

    "github.com/bozso/gomma/data"
)

const maxIW = 3

type IW struct {
    data.ComplexWithPar
    TOPSPar path.ValidFile
}

type IWs [maxIW]IW

func (iw IW) Tabline() (s string) {
    s = fmt.Sprintf("%s %s %s\n", iw.DatFile, iw.ParFile, iw.TOPSPar)
    return
}

func (iw IW) Move(dir path.Dir) (miw IW, err error) {
    miw = iw
    
    if miw.ComplexWithPar.File, err = iw.ComplexWithPar.Move(dir); err != nil {
        return
    }
    
    miw.TOPSPar, err = iw.TOPSPar.Move(dir)
    return
}
