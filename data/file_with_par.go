package data

import (
    "github.com/bozso/gotoolbox/path"
)

type Parameter struct {
    ParFile path.ValidFile `json:"parfile"`    
}

type FileWithPar struct {
    Parameter
    File
}

func (f *FileWithPar) Set(s string) (err error) {
    return LoadJson(s, f)
}
