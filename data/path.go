package data

import (
    "github.com/bozso/gomma/common"    
    "github.com/bozso/gotoolbox/path"
)


type Path struct {
    DatFile path.File
}

func New(file path.File) (p Path) {
    p.DatFile = file
    return
}

func (p Path) Load(ra common.RngAzi, dtype Type) (f File, err error) {
    f.DatFile, err = p.DatFile.ToValid()
    if err != nil {
        return
    }
    
    f.Ra, f.Dtype = ra, dtype
    return
}
