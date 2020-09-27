package data

import (
    "github.com/bozso/gotoolbox/path"
    "github.com/bozso/gomma/common"
)

type Typer interface {
    Type() Type
}

type DataFile interface {
    Typer
    common.Pather
    common.Dims
}

type Saver interface {
    Save() error
    SaveWithPath(file path.File) (err error)
}
