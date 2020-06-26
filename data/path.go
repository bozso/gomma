package data

import (
    "github.com/bozso/gomma/common"    
    "github.com/bozso/gotoolbox/path"
)


type Path struct {
    DatFile path.File
}

func New(file path.Pather) (p Path) {
    p.DatFile = file.AsPath().ToFile()
    return
}

func (d File) WithShape(p Path) (f File, err error) {
    f, err = d.WithShapeDType(p, d.Dtype)
    return
}

func (d File) WithShapeDType(p Path, dtype Type) (f File, err error) {
    if dtype == Unknown {
        dtype = d.Dtype
    }
    
    f, err = p.Load(d.Ra, dtype)
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
