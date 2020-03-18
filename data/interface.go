package data

import (
    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gomma/common"
)

type (
    Pather interface {
        DataPath() path.ValidFile
    }
    
    Typer interface {
        DataType() Type
    }
    
    Data interface {
        Typer
        Pather
        common.Dims
    }
    
    Saver interface {
        Save() error
        SaveWithPath(file path.File) (err error)
    }
)
